// Refresh offline CLI fixtures using the API serializer and pinned SDK examples.
'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const [sdkPath, formatterPath, flag] = process.argv.slice(2);
assert(sdkPath && formatterPath && (!flag || flag === '--check'), 'Usage: node scripts/update_response_fixtures.cjs SDK_ROOT RESPONSE_FORMATTING_JS [--check]');
const { formatResponse } = require(path.resolve(formatterPath));
const examples = JSON.parse(fs.readFileSync(path.join(sdkPath, 'testdata/source-examples.json'), 'utf8'));
function cases(body, content) {
  const rows = [];
  for (const accept of ['application/json', 'text/plain'])
    for (const respondWith of content ? [null, 'markdown'] : [null])
      rows.push({ accept, respondWith, response: formatResponse(body, { accept, representation: respondWith ?? 'original', content }) });
  return rows;
}
const fixture = { sources: {}, batch: {} };
fixture.redaction = {
  secrets: ['fd_credentialDummy_key', 'knownCredentialDummy_&<>*`-secret', 'multilineCredentialDummy\r\nsecond_line'],
};
fixture.redaction.responses = cases({
  source: 'google_search', status: 'done', results: [{
    searchParameters: { q: fixture.redaction.secrets[0] },
    echoes: fixture.redaction.secrets.slice(1),
    unknownKey: 'fd_unknownCredentialDummy-key',
    link: 'https://example.com/article?q=provenance',
    literal: 'Unrelated *_[] & <>',
  }],
}, true).filter(row => row.respondWith === 'markdown');
for (const [source, example] of Object.entries(examples)) {
  const stored = {
    id: 'job_offline', source, status: 'completed',
    results: source === 'youtube_download' ? [{ file_url: 'https://customer.example/videos/example.mp4', downloaded_bytes: 1024, file_size: 1024, md5: '0123456789abcdef0123456789abcdef', videoId: 'dQw4w9WgXcQ' }] : example.response.results,
    ...(source === 'youtube_download' ? {} : { usage: example.response.usage }),
    completed_at: '2026-10-23T00:00:00Z', expires_at: '2026-10-24T00:00:00Z',
  };
  fixture.sources[source] = {
    query: example.request,
    ...(source === 'youtube_download' ? {} : { realtime: cases(example.response, true) }),
    stored: cases(stored, true),
    submission: cases({ id: 'job_offline', source, status: 'pending', usage: { credits: 0.02 } }, false),
    status: cases({ job: { id: 'job_offline', source, status: 'completed', params: example.request } }, false),
  };
}
const queries = [examples.google_search.request, examples.youtube_transcript.request];
fixture.batch = {
  queries,
  responses: cases({ queries: queries.map((query, i) => ({ id: i ? `job_offline${i}` : 'job_offline', source: query.source, status: 'pending' })), usage: { credits: 0.04 } }, false),
};
const file = path.join(__dirname, '../internal/app/testdata/response-formats.json');
if (flag) assert.deepEqual(JSON.parse(fs.readFileSync(file, 'utf8')), fixture, 'CLI response fixtures are stale.');
else { fs.mkdirSync(path.dirname(file), { recursive: true }); fs.writeFileSync(file, JSON.stringify(fixture, null, 2) + '\n'); }
console.log(flag ? 'CLI fixtures match the API serializer and SDK examples.' : 'Updated CLI response fixtures.');
