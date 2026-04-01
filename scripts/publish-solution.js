#!/usr/bin/env node
//
// Publish or update a Litmus solution entry in Contentful.
//
// This script upserts a "solution" content type entry in Contentful based on
// a JSON config file (default: solution.json). It is used both
// locally and in CI (via the publish-solution GitHub Action).
//
// Behavior:
//   - CREATE: If no entry with the config's slug exists, a new entry is created
//     with all fields from the config. Linked entries (logo, tags, etc.) are
//     optional — missing ones produce a warning but don't block creation.
//   - UPDATE (default): If an entry already exists, only `latestVersion` and
//     `releaseDate` are patched. All other fields are left untouched.
//   - UPDATE (--force-update): All fields from the config are applied to the
//     existing entry, except `slug` which is immutable.
//   - IDEMPOTENT: If the entry is already at the requested version, the script
//     exits early with no changes (unless --force-update is set).
//
// Usage:
//   npm run solution:publish -- --version 0.1.0
//   npm run solution:publish -- --version v0.1.0 --force-update
//   npm run solution:publish -- --version 0.1.0 --no-publish
//   npm run solution:publish -- --version 0.1.0 --config path/to/config.json
//   npm run solution:publish -- --version 0.1.0 --environment development
//
// Environment variables (loaded from .env if present):
//   CONTENTFUL_SPACE_ID           Contentful space ID (required)
//   CONTENTFUL_MANAGEMENT_TOKEN   Contentful Management API token (required)
//   CONTENTFUL_ENVIRONMENT        Contentful environment (default: master)

const { createClient } = require('contentful-management');
const fs = require('fs');
const path = require('path');

// ---------------------------------------------------------------------------
// Parse CLI arguments
// ---------------------------------------------------------------------------
const args = process.argv.slice(2);
let version = '';
let forceUpdate = false;
let shouldPublish = true;
let configPath = path.resolve('solution.json');
let environment = '';

for (let i = 0; i < args.length; i++) {
  switch (args[i]) {
    case '--version':
      version = args[++i];
      break;
    case '--force-update':
      forceUpdate = true;
      break;
    case '--no-publish':
      shouldPublish = false;
      break;
    case '--config':
      configPath = path.resolve(args[++i]);
      break;
    case '--environment':
      environment = args[++i];
      break;
    case '--help':
      console.log(`Usage: npm run solution:publish -- --version <version> [options]

Options:
  --version <version>       Release version (required, v-prefix stripped automatically)
  --force-update            Apply all config fields on update (default: only version + date)
  --no-publish              Save as draft, do not publish
  --config <path>           Path to config JSON (default: solution.json)
  --environment <env>       Contentful environment (default: CONTENTFUL_ENVIRONMENT or master)

Environment variables:
  CONTENTFUL_SPACE_ID           Contentful space ID (required)
  CONTENTFUL_MANAGEMENT_TOKEN   Contentful Management API token (required)
  CONTENTFUL_ENVIRONMENT        Contentful environment (default: master)`);
      process.exit(0);
  }
}

// ---------------------------------------------------------------------------
// Load .env (does not override existing env vars)
// ---------------------------------------------------------------------------
const envPath = path.resolve('.env');
if (fs.existsSync(envPath)) {
  for (const line of fs.readFileSync(envPath, 'utf8').split('\n')) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;
    const eqIdx = trimmed.indexOf('=');
    if (eqIdx === -1) continue;
    const key = trimmed.slice(0, eqIdx);
    const val = trimmed.slice(eqIdx + 1).replace(/^(['"])(.*)\1$/, '$2');
    if (!process.env[key]) process.env[key] = val;
  }
}

// ---------------------------------------------------------------------------
// Validate required inputs
// ---------------------------------------------------------------------------
const spaceId = process.env.CONTENTFUL_SPACE_ID;
const managementToken = process.env.CONTENTFUL_MANAGEMENT_TOKEN;
const envId = environment || process.env.CONTENTFUL_ENVIRONMENT || 'master';

if (!version) {
  console.error('Error: --version is required');
  process.exit(1);
}
if (!spaceId) {
  console.error('Error: CONTENTFUL_SPACE_ID is not set');
  process.exit(1);
}
if (!managementToken) {
  console.error('Error: CONTENTFUL_MANAGEMENT_TOKEN is not set');
  process.exit(1);
}

// Strip leading "v" (e.g. v0.1.0 -> 0.1.0) and validate format.
version = version.replace(/^v/, '');
const versionPattern = /^([a-zA-Z]+-)?\d+(\.\d+)*x?$/;
if (!versionPattern.test(version)) {
  console.error(`Error: version "${version}" does not match Contentful pattern`);
  process.exit(1);
}

// ---------------------------------------------------------------------------
// Read and validate config
// ---------------------------------------------------------------------------
if (!fs.existsSync(configPath)) {
  console.error(`Error: config file not found: ${configPath}`);
  process.exit(1);
}

let config;
try {
  config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
} catch (e) {
  console.error(`Error: failed to parse config: ${e.message}`);
  process.exit(1);
}

// These fields are required by the Contentful solution content type and must
// be present in the config for a successful create or force-update.
const requiredConfigFields = [
  'name',
  'slug',
  'excerpt',
  'description',
  'vendor',
  'vendorUrl',
  'documentationUrl',
  'downloadingUrl',
];
const missingFields = requiredConfigFields.filter((f) => !config[f]);
if (missingFields.length > 0) {
  console.error(`Error: config is missing required fields: ${missingFields.join(', ')}`);
  process.exit(1);
}

// ---------------------------------------------------------------------------
// Contentful field helpers
//
// All Contentful fields are localized — even single-locale spaces wrap values
// in a { "en-US": value } object. Linked entries use the CMA Link format:
// { sys: { type: "Link", linkType: "Entry", id: "<entry-id>" } }
// ---------------------------------------------------------------------------
const LOCALE = 'en-US';
const CONTENT_TYPE = 'solution';

/** Wrap a Contentful entry ID in a CMA Link object. */
function entryLink(id) {
  return { sys: { type: 'Link', linkType: 'Entry', id } };
}

/** Wrap a value in a locale envelope. */
function localized(value) {
  return { [LOCALE]: value };
}

/** Build a localized single-entry link, or undefined if no ID provided. */
function localizedLink(id) {
  return id ? { [LOCALE]: entryLink(id) } : undefined;
}

/** Build a localized array of entry links, or undefined if empty/missing. */
function localizedLinkArray(ids) {
  return ids && ids.length > 0 ? { [LOCALE]: ids.map(entryLink) } : undefined;
}

/**
 * Build the full Contentful fields object from config + derived values.
 * Used on create and on force-update.
 */
function buildAllFields(config, version) {
  const releaseDate = new Date().toISOString().split('T')[0];

  // Scalar / array fields populated directly from config
  const fields = {
    name: localized(config.name),
    slug: localized(config.slug),
    excerpt: localized(config.excerpt),
    description: localized(config.description),
    vendor: localized(config.vendor),
    vendorUrl: localized(config.vendorUrl),
    documentationUrl: localized(config.documentationUrl),
    downloadingUrl: localized(config.downloadingUrl),
    productCategory: localized(config.productCategory || ['Litmus Edge']),
    litmusEdgeVersions: localized(config.litmusEdgeVersions || []),
    licenseType: localized(config.licenseType || 'Open Source'),
    userSupport: localized(config.userSupport || 'Community'),
    verified: localized(config.verified !== undefined ? config.verified : true),
    featured: localized(config.featured || false),
    downloads: localized(config.downloads || {}),
    // Derived from the --version flag and current date
    latestVersion: localized(version),
    releaseDate: localized(releaseDate),
  };

  // Linked entries — each value in the config is a Contentful entry ID (or
  // array of IDs). They are optional: missing ones produce a warning for
  // fields the Contentful schema marks as required (logo, featuredImage,
  // screenshots, tags) and are silently skipped for truly optional ones.
  const linkFields = {
    logo: () => localizedLink(config.logo),
    featuredImage: () => localizedLink(config.featuredImage),
    thumbnailImage: () => localizedLink(config.thumbnailImage),
    screenshots: () => localizedLinkArray(config.screenshots),
    tags: () => localizedLinkArray(config.tags),
    relatedSolutions: () => localizedLinkArray(config.relatedSolutions),
    section: () => localizedLink(config.section),
  };

  const requiredLinks = ['logo', 'featuredImage', 'screenshots', 'tags'];

  for (const [key, builder] of Object.entries(linkFields)) {
    const value = builder();
    if (value) {
      fields[key] = value;
    } else if (requiredLinks.includes(key)) {
      console.warn(`Warning: "${key}" not provided in config — entry will be incomplete`);
    }
  }

  return fields;
}

/**
 * Build the minimal fields object for a routine version bump.
 * Only latestVersion and releaseDate change on a normal update.
 */
function buildUpdateFields(version) {
  const releaseDate = new Date().toISOString().split('T')[0];
  return {
    latestVersion: localized(version),
    releaseDate: localized(releaseDate),
  };
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------
async function main() {
  console.log(`Space: ${spaceId} | Environment: ${envId}`);
  console.log(`Config: ${configPath}`);
  console.log(`Solution: ${config.slug} | Version: ${version}`);
  console.log(`Mode: ${forceUpdate ? 'force-update' : 'auto'} | Publish: ${shouldPublish}`);
  console.log('');

  const client = createClient({ accessToken: managementToken });
  const params = { spaceId, environmentId: envId };

  // Look up existing entry by slug
  console.log(`Looking up solution with slug "${config.slug}"...`);
  const entries = await client.entry.getMany({
    ...params,
    query: {
      content_type: CONTENT_TYPE,
      'fields.slug': config.slug,
      limit: 1,
    },
  });

  if (entries.items.length === 0) {
    // --- Create new entry ---
    console.log('No existing entry found — creating new solution...');
    const fields = buildAllFields(config, version);
    const entry = await client.entry.create({ ...params, contentTypeId: CONTENT_TYPE }, { fields });
    console.log(`Created entry: ${entry.sys.id}`);

    console.log('Saved as draft — review in Contentful before publishing.');
  } else {
    // --- Update existing entry ---
    const entry = entries.items[0];
    const currentVersion = entry.fields.latestVersion?.[LOCALE];
    console.log(`Found entry: ${entry.sys.id} (current version: ${currentVersion})`);

    // Skip if already at the target version (unless force-updating all fields)
    if (currentVersion === version && !forceUpdate) {
      console.log(`Already at version ${version} — nothing to update.`);
      return;
    }

    if (forceUpdate) {
      // Replace all fields from config (slug is preserved from the existing entry)
      console.log('Force update — applying all fields from config...');
      const fields = buildAllFields(config, version);
      delete fields.slug;
      entry.fields = { slug: entry.fields.slug, ...fields };
    } else {
      // Normal release bump — only version and date
      console.log(`Updating version: ${currentVersion} -> ${version}`);
      const fields = buildUpdateFields(version);
      Object.assign(entry.fields, fields);
    }

    // CMA requires the current sys.version for optimistic locking
    const updated = await client.entry.update(
      { ...params, entryId: entry.sys.id },
      { sys: entry.sys, fields: entry.fields }
    );
    console.log('Entry updated.');

    if (shouldPublish) {
      await client.entry.publish({ ...params, entryId: updated.sys.id }, { sys: updated.sys });
      console.log('Published.');
    } else {
      console.log('Saved as draft (--no-publish).');
    }
  }

  console.log('Done.');
}

main().catch((err) => {
  console.error('Error:', err.message || err);
  if (err.details?.errors) {
    for (const e of err.details.errors) {
      console.error(`  - ${e.name}: ${e.details || e.value || JSON.stringify(e)}`);
    }
  }
  process.exit(1);
});
