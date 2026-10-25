"""Read a complete MCP result without retrying the original tool call."""
import json
from datetime import timedelta

MAX_UNITS = 4 * 1024 * 1024
MAX_READS = 4096


def response_body(result):
    if result.isError:
        raise RuntimeError("MCP tool reported an error; no submission was retried.")
    if result.structuredContent is not None:
        return result.structuredContent
    texts = [item.text for item in result.content if item.type == "text"]
    # Formatted chunks keep metadata and readable content in separate text blocks
    # so hosts that omit structuredContent can still read the whole document.
    if len(texts) >= 2:
        try:
            metadata = json.loads(texts[0])
        except ValueError:
            metadata = None
        if (isinstance(metadata, dict) and "result_id" in metadata and "next_offset" in metadata
                and metadata.get("encoding") in ("markdown", "text") and "chunk" not in metadata):
            return {**metadata, "chunk": "".join(texts[1:])}
    text = "".join(texts)
    try:
        return json.loads(text)
    except ValueError:
        return {"content": text, "format": "text"}


async def read_complete(session, result):
    body = response_body(result)
    if not isinstance(body, dict) or "result_id" not in body:
        return body
    result_id = body.get("result_id")
    encoding = body.get("encoding", "json")
    total = body.get("total_characters")
    if (not isinstance(result_id, str) or not result_id or encoding not in ("json", "markdown", "text")
            or type(total) is not int or not 0 <= total <= MAX_UNITS):
        raise RuntimeError("Invalid result continuation metadata.")
    parts = []
    offset = 0
    for _ in range(MAX_READS):
        chunk = body.get("chunk")
        if (body.get("result_id") != result_id or body.get("encoding", "json") != encoding
                or body.get("total_characters") != total or type(body.get("offset")) is not int
                or body["offset"] != offset or not isinstance(chunk, str)):
            raise RuntimeError("Result continuation changed or has missing chunks.")
        offset += len(chunk.encode("utf-16-le", "surrogatepass")) // 2
        if offset > total or "next_offset" not in body:
            raise RuntimeError("Invalid result continuation offset.")
        parts.append(chunk)
        next_offset = body["next_offset"]
        if next_offset is None:
            if offset != total:
                raise RuntimeError("Incomplete result; no submission was retried.")
            # JavaScript offsets can split a surrogate pair between chunks.
            text = "".join(parts).encode("utf-16-le", "surrogatepass").decode("utf-16-le")
            return json.loads(text) if encoding == "json" else {"content": text, "format": encoding}
        if type(next_offset) is not int or next_offset != offset or not chunk or offset >= total:
            raise RuntimeError("Invalid result continuation offset.")
        result = await session.call_tool("get_result", arguments={"result_id": result_id, "offset": next_offset},
                                         read_timeout_seconds=timedelta(seconds=120))
        body = response_body(result)
        if not isinstance(body, dict):
            raise RuntimeError("Invalid result continuation response.")
    raise RuntimeError("Too many result chunks; no submission was retried.")


async def call_complete(session, name, arguments):
    """Python examples request structured JSON when supported by the server."""
    arguments = dict(arguments)
    cursor = None
    for _ in range(100):
        page = await session.list_tools(cursor=cursor)
        tool = next((tool for tool in page.tools if tool.name == name), None)
        if tool is not None:
            formats = tool.inputSchema.get("properties", {}).get("response_format", {}).get("enum", [])
            if "json" in formats:
                arguments.setdefault("response_format", "json")
            break
        cursor = page.nextCursor
        if cursor is None:
            raise RuntimeError("Requested tool is unavailable.")
    else:
        raise RuntimeError("Tool discovery did not finish.")
    result = await session.call_tool(name, arguments=arguments,
                                    read_timeout_seconds=timedelta(seconds=120))
    return await read_complete(session, result)
