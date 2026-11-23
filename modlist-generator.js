#!/usr/bin/env node

const KB256 = 256 * 1024

const fs = require('fs');
const path = require('path');
const { execFile } = require('node:child_process');
const { promisify } = require('node:util');

const execFileAsync = promisify(execFile);

const Success = (value) => ({ isSuccess: true, value });
const Failure = (error) => ({ isFailure: true, value: error });

const STD = {}
const TASK = {}

STD.map = (func) => (data) =>
  data.isFailure
    ? data
    : STD.try(func)(data.value);

STD.try = (func) => (value) => {
  try {
    return Success(func(value));
  } catch (error) {
    return Failure(error);
  }
};

STD.pipe = (value, ...funcs) =>
  funcs.reduce(
    (result, func) => STD.map(func)(result),
    Success(value)
  );

STD.fold = (failureHandler, successHandler) => (value) =>
  value.isFailure
    ? failureHandler(value.value)
    : successHandler(value.value);

STD.partition = (data) =>
  Object.entries(data).reduce(
    (acc, [key, data]) => {
      if (data.isFailure) {
        acc.failed[key] = data.value;
      } else {
        acc.success[key] = data.value;
      }

      return acc;
    },
    { success: {}, failed: {} }
  )

TASK.map = (func) => (task) =>
  () =>
    task().then((data) =>
      data.isFailure
        ? data
        : STD.try(func)(data.value)
    );

TASK.chain = (func) => (task) =>
  () =>
    task().then((data) =>
      data.isFailure
        ? data
        : func(data.value)()
    );

TASK.Success = (value) => () =>
  Promise.resolve(Success(value));

TASK.Failure = (error) => () =>
  Promise.resolve(Failure(error));

TASK.tryCatch = (func) => (...args) => () =>
  Promise.resolve()
    .then(() => func(...args))
    .then(Success)
    .catch(Failure);

TASK.fold = (failureHandler, successHandler) => (task) =>
  task().then((data) =>
    data.isFailure
      ? failureHandler(data.value)
      : successHandler(data.value)
  );

TASK.from = (value) => () =>
  Promise.resolve(value);

TASK.pipe = (task, ...funcs) =>
  funcs.reduce(
    (result, func) => func(result),
    task
  );

const escapeRegex = (text) =>
  text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

const extractValue = (rawString, key) => {
  const regex = new RegExp(
    `"${escapeRegex(key)}"\\s*:\\s*"((?:\\\\.|[^"\\\\])*)"`
  );

  const got = rawString.match(regex);

  return !got
    ? Failure(new Error(`Unable to extract field "${key}"`))
    : Success(got[1]);
};

const extractValuesObject = (...keys) => (rawString) =>
  Object.fromEntries(
    keys.map((key) => [
      key,
      extractValue(rawString, key)
    ])
  );

/**
 * Vintage Story Mod List Generator
 *
 * Extract mod metadata from .zip archives and render a formatted summary.
 *
 * Usage:
 *   modlist-generator.js [options]
 *
 * Options:
 *   -m, --markdown       Output as a Markdown table with summary section.
 *                        (default: plain text one-per-line)
 *   -h, --help           Show this help message and exit.
 *
 * Notes:
 *   Reads all .zip files in the current working directory.  Each archive is
 *   opened for its modinfo.json entry; if the file is missing or malformed
 *   the mod lands in a "Failed Entries" block instead of being silently
 *   dropped.
 */

const processModInfo = (buffer) =>
  STD.pipe(
    buffer,
    (buffer) => buffer.toString(),
    (text) => text.replace(/\ufeff/g, ''),
    (text) => text.replace(/\r\n/g, '\n'),
    (text) => text.replace(/,\s+\}/g, ''),
    (text) => text.replace(/"([A-Z])([^"]*)":/g, (_, p1, p2) => `"${p1.toLowerCase()}${p2}":`),
    extractValuesObject('name', 'version')
  );

const extractFromZip = TASK.tryCatch((zipPath) =>
  execFileAsync(
    'unzip',
    ['-p', zipPath, 'modinfo.json'],
    { maxBuffer: KB256 }
  ).then(({ stdout }) => stdout)
);

const extractModInfo = (zipPath) =>
  TASK.pipe(
    extractFromZip(zipPath),
    TASK.chain((buffer) =>
      TASK.from(processModInfo(buffer))
    )
  );

const concatLine = (line, lines) => 
  !lines || lines.length === 0 ? 
    line : line ? 
      `${line}\n${concatLine(lines[0], lines.slice(1))}` :
      `${concatLine(lines[0], lines.slice(1))}`;

const mdRender = (results, failed ) => {

  let err
  let output

  try {
    output = [].concat(
      `# Mod List\n\n| Name | Version |\n| ------ | --------- |`,
      results
        .map(mod => `| \`${mod.name}\` | \`${mod.version}\` |`),
      `\n## Summary\n\nTotal entries: **${results.length}**`,
      failed.length > 0 ? ` (${failed.length} missing metadata)\n\n## Failed Entries\n` : '',
      failed.map(failedLine => `* \`${failedLine.zipName}\` - ${Object.values(failedLine.fields).join(', ')}`),
      ''
    )
  } catch (error) {
    err = error;
  }

  return [concatLine(null, output), err]
}

const rawRender = (results, failed ) => {

  let err
  let output

  try {
    output = [].concat(
      results.map(mod => `${mod.name} ${mod.version}`),
      failed.map(failedLine => `# ERROR \`${failedLine.zipName}\`: ${Object.values(failedLine.fields).join(', ')}`),
      ''
    )
  } catch (error) {
    err = error;
  }

  return [concatLine(null, output), err]
}

async function main() {
  const modsDir = process.cwd();
	const showHelp   = process.argv.includes('-h') || process.argv.includes('--help');
	if (showHelp) {
		process.stdout.write(`Usage:
  modlist-generator.js [options]

Options:
  -m, --markdown       Output as a Markdown table with summary section.
                       (default: plain text one-per-line)
  -h, --help           Show this help message and exit.

Notes:
  Reads all .zip files in the current working directory. Each archive is
  opened for its modinfo.json entry; if the file is missing or malformed
  the mod lands in a "Failed Entries" block instead of being silently dropped.
`);
		process.exit(0);
	}

	const parameters = {md: process.argv.includes('--markdown')};

  if (!fs.existsSync(modsDir)) {
    console.error('Mods directory not found:', modsDir);
    process.exit(1);
  }

  // Get all .zip files sorted alphabetically
  const zipFiles = fs.readdirSync(modsDir)
    .filter(f => f.endsWith('.zip'))
    .sort();

  if (zipFiles.length === 0) {
    console.error('No .zip files found in Mods directory.');
    process.exit(0);
  }

  const results = [];
  const failed = [];

  for (const zipFile of zipFiles) {
    await TASK.fold(
      (error) => failed.push({ zipFile, error }),
      (modInfo) =>   {
        const { success, failed: fieldErrors } = STD.partition(modInfo);

        !!Object.keys(fieldErrors).length ?
          failed.push({
            zipName: zipFile,
            fields: fieldErrors
          }) :
          results.push({
            zipName: zipFile,
            ...success
          });
        }   
    )(
      extractModInfo(path.join(modsDir, zipFile))
    );
  }

  results.sort((a, b) => {
    return a.name && b.name ? a.name.localeCompare(b.name) : 0;
  })
  failed.sort((a, b) => a.zipName.localeCompare(b.zipName))

  const [output, err] = parameters.md ? mdRender(results, failed) : rawRender(results, failed);
  if (err) {
    console.error('Error generating mod list:', err);
    process.exit(1);
  }
  process.stdout.write(output);
  process.exit(0);
}


main().catch(err => {
  console.error('Fatal error:', err);
  process.exit(1);
});
