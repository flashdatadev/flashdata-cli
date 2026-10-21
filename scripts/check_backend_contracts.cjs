// Optional development check against a local backend checkout. No server,
// registration, paid query, database or customer-cloud operation is started.
'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { execFileSync } = require('node:child_process');

const root = path.resolve(__dirname, '..');
if (!process.argv[2]) throw new Error('Usage: node scripts/check_backend_contracts.cjs /path/to/backend [/path/to/sdk]');
const backend = path.resolve(process.argv[2]);
const { SOURCE_CATALOG } = require(path.join(backend, 'modules/catalog/sourceCatalog'));
const { validateRegistration } = require(path.join(backend, 'modules/agentConnections/policy'));
const { normalizeOptions } = require(path.join(backend, 'modules/bulk/bulkJobService'));
const { normalizeYouTubeSearchOptions } = require(path.join(backend, 'modules/youtube/youtubeSearchOptions'));
const sdkRoot = process.argv[3] ? path.resolve(process.argv[3]) : execFileSync('go', ['list', '-m', '-f', '{{.Dir}}', 'github.com/flashdatadev/flashdata-go'], { cwd: root, encoding: 'utf8', env: { ...process.env, GOWORK: 'off' } }).trim();
const schemas = JSON.parse(fs.readFileSync(path.join(sdkRoot, 'internal/contract/schemas.json'), 'utf8'));
assert.deepEqual(Object.keys(schemas).sort(), SOURCE_CATALOG.map(source => source.id).sort());
for (const source of SOURCE_CATALOG) {
  const schema = schemas[source.id];
  const properties = { ...schema.properties, ...schema.properties.params?.properties };
  for (const [key, definition] of Object.entries(source.inputSchema.properties)) {
    let sdkDefinition = properties[key];
    if (key === 'filters' && ['youtube_search', 'youtube_search_extended'].includes(source.id)) {
      // The Data API explicitly treats filters:null as omitted, although the
      // public catalog describes only the populated object shape.
      assert.deepEqual(normalizeYouTubeSearchOptions({ filters: null }), {});
      assert.deepEqual(sdkDefinition.type, ['object', 'null']);
      sdkDefinition = { ...sdkDefinition, type: 'object' };
    }
    assert.deepEqual(sdkDefinition, definition, `${source.id}.${key} changed in backend`);
  }
  const required = new Set([...schema.required, ...(schema.properties.params?.required || [])]);
  for (const key of source.inputSchema.required) assert.ok(required.has(key), `${source.id}.${key} must be required`);
}
validateRegistration({ name: 'Offline CLI contract check', audience: 'coding', client: 'flashdata-cli', transport: 'cli', scopes: ['google_search', 'youtube_metadata'] });
for (const source of ['youtube_metadata', 'youtube_transcript', 'youtube_captions', 'youtube_trainability']) normalizeOptions(source, {});
normalizeOptions('youtube_download', { download_type: 'audio', output_format: 'mp3', storage: { storage_type: 's3', storage_url: 'offline-bucket', access_key: 'offline-access', secret_key: 'offline-secret' } });
const create = JSON.parse(fs.readFileSync(path.join(root, 'examples/bulk-create.json'), 'utf8'));
normalizeOptions(create.source, create.options);
console.log(`Backend contract check passed: ${SOURCE_CATALOG.length} source schemas, five bulk sources, registration and the manual-create example.`);
