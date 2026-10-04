import { test, expect } from 'e2e';
import { readFile, stat, writeFile, access, symlink } from 'node:fs/promises';
import { join } from 'node:path';
import { workspace, fileBytes } from './support';

let w: Awaited<ReturnType<typeof workspace>>;
test.beforeEach(async () => { w = await workspace(); });
test.afterEach(async () => { await w?.close(); });

async function success(args: string[]) {
  const result = await w.run(args);
  expect(result.code, result.stderr).toBe(0);
  expect(result.stderr).toBe('');
  return result;
}
async function json(args: string[]) { return JSON.parse((await success(['--json', ...args])).stdout); }
async function failure(args: string[], code: number, message: string) {
  const result = await w.run(args);
  expect(result.code, `${result.stdout}\n${result.stderr}`).toBe(code);
  expect(result.stdout).toBe('');
  expect(result.stderr).toContain(message);
}
function last(path: string) { return w.requests.filter((r) => r.path === path).at(-1); }
async function absent(path: string) { expect(await access(path).then(() => false, () => true)).toBe(true); }

// Existing Go tests call Execute directly. These tests own compiled-process wiring,
// real exit codes, authentication cookies, and durable output from complete flows.
test('config creation persists private credentials, refuses overwrite, and supports effective overrides', async () => {
  const path = join(w.dir, 'new', 'config.yaml');
  const args = ['--json', 'config', 'init', '--username', 'synthetic-user', '--password-stdin', '--base-url', w.baseURL, '--origin-host', 'localhost'];
  const result = await w.run(args, { config: path, stdin: ' synthetic-password \r\n' });
  expect(result.code, result.stderr).toBe(0);
  expect(JSON.parse(result.stdout)).toEqual({ path });
  expect((await stat(path)).mode & 0o777).toBe(0o600);
  expect((await stat(join(w.dir, 'new'))).mode & 0o777).toBe(0o700);
  const saved = await readFile(path, 'utf8');
  expect(saved).toContain(' synthetic-password ');
  const shown = await w.run(['--json', 'config', 'show'], { config: path });
  expect(shown.code, shown.stderr).toBe(0);
  expect(JSON.parse(shown.stdout)).toEqual({ path, base_url: w.baseURL, origin_host: 'localhost', username: 'redacted', password: 'redacted' });
  expect((await w.run(['--json', 'login'], { config: path })).code).toBe(0);
  const duplicate = await w.run(args, { config: path, stdin: 'replacement' });
  expect(duplicate.code).toBe(1);
  expect(duplicate.stderr).toContain('already exists');
  expect(await readFile(path, 'utf8')).toBe(saved);
  const forced = await w.run([...args, '--force'], { config: path, stdin: 'replacement' });
  expect(forced.code).toBe(0);
  expect(await readFile(path, 'utf8')).toContain('password: replacement');
  const override = await w.run(['--json', '--base-url', w.baseURL, '--origin-host', 'flag.example', 'config', 'show'], { env: { CLASSREACH_BASE_URL: 'https://env.example', CLASSREACH_ORIGIN_HOST: 'env.example', CLASSREACH_USERNAME: 'env-user' } });
  expect(override.code).toBe(0);
  expect(JSON.parse(override.stdout).origin_host).toBe('flag.example');
  expect(JSON.parse(override.stdout).base_url).toBe(w.baseURL);
  expect(override.stdout).not.toContain('env-user');
});

test('login and doctor distinguish missing credentials, denied auth, challenges, and absent sessions', async () => {
  expect((await json(['login'])).ok).toBe(true);
  expect((await json(['doctor'])).ok).toBe(true);
  for (const [mode, message] of [['denied', 'login failed'], ['missing-token', 'anti-forgery token'], ['no-session', 'no authenticated session']]) {
    w.state.login = mode;
    await failure(['--json', 'login'], 1, message);
  }
  w.state.login = 'ok';
  const empty = join(w.dir, 'empty.yaml');
  await writeFile(empty, `base_url: ${w.baseURL}\norigin_host: localhost\n`);
  const before = w.requests.length;
  const missing = await w.run(['overview'], { config: empty });
  expect(missing.code).toBe(1);
  expect(missing.stderr).toContain('CLASSREACH_USERNAME');
  expect(w.requests.length).toBe(before);
  await writeFile(empty, `base_url: ${w.baseURL}\norigin_host: localhost\nusername: synthetic-user\n`);
  expect((await w.run(['overview'], { config: empty })).stderr).toContain('CLASSREACH_PASSWORD');
  await writeFile(empty, 'base_url: [broken');
  expect((await w.run(['overview'], { config: empty })).stderr).toContain('parse config');
  w.state.response = 'empty-view';
  await failure(['--json', 'doctor'], 1, 'missing UserInfos');
});

test('guardian flow resolves students, filters sections, retrieves grades, and renders announcements', async () => {
  const overview = await json(['overview', '--week', '2026-10-05']);
  expect(overview.week).toBe('2026-10-05');
  expect(overview.quick_view.UserInfos.map((s: any) => s.UserID)).toEqual(['student-1', 'student-2']);
  expect(last('/Home/GetQuickView').query.get('weekDate')).toBe('2026-10-05T00:00:00');
  expect((await json(['students', 'ls'])).map((s: any) => s.Name)).toEqual(['Alex Example', 'Casey Example']);
  expect((await json(['students', 'get', 'student-1'])).Name).toBe('Alex Example');
  expect((await success(['students', 'get', 'student-1'])).stdout).toContain('summary_url');
  await failure(['students', 'get', 'unknown'], 1, 'not visible');
  const courses = await json(['courses', 'list', '--student', 'student-1']);
  expect(courses.map((c: any) => c.section_id)).toEqual(['section-1']);
  expect(await json(['courses', 'list', '--student', 'student-2'])).toEqual([]);
  expect((await json(['courses', 'get', 'section-1'])).course_name).toBe('Science');
  expect((await success(['courses', 'get', 'section-1'])).stdout).toContain('Science');
  await failure(['courses', 'get', 'unknown'], 1, 'not visible');
  expect((await json(['grades', 'list', '--student', 'student-1']))[0].numeric_grade).toBe(94.5);
  expect((await success(['grades', 'list'])).stdout).toContain('Science');
  expect((await success(['overview', '--week', '2026-10-05'])).stdout).toContain('Students: 2');
  expect((await success(['students', 'list'])).stdout).toContain('Alex Example');
  expect((await success(['courses', 'list'])).stdout).toContain('Science');
  expect((await json(['announcements', 'ls']))[0].Heading).toBe('School fair');
  const announcements = (await success(['announcements', 'list'])).stdout;
  expect(announcements).toContain('Bring books & games.');
  expect(announcements).not.toContain('<b>');
});

test('academic flow reads escaped HTML models, resolves assignments, and maps attendance markings', async () => {
  const flags = ['--student', 'student-1', '--section', 'section-1'];
  expect((await json(['assignments', 'list', ...flags]))[0].Assignment.Name).toBe("Earth's orbit");
  expect((await json(['assignments', 'get', 'assignment-1', ...flags])).Assignment.ID).toBe('assignment-1');
  expect((await success(['assignments', 'list', ...flags])).stdout).toContain("Earth's orbit");
  expect((await success(['assignments', 'get', 'assignment-1', ...flags])).stdout).toContain('Homework');
  await failure(['assignments', 'get', 'unknown', ...flags], 1, 'was not found');
  expect((await json(['attendance', 'list', ...flags])).studentAttendance.Attendance[0].ID).toBe('attendance-1');
  expect((await success(['attendance', 'list', ...flags])).stdout).toContain('Present');
  w.state.response = 'html';
  await failure(['assignments', 'list', ...flags], 1, 'marker');
  await failure(['attendance', 'list', ...flags], 1, 'marker');
});

test('calendar and notifications preserve requested dates, term, and provider numbers', async () => {
  const args = ['calendar', 'list', '--start', '2026-10-01', '--end', '2026-10-31'];
  expect((await json(args))[0].CalendarEvent.Name).toBe('School fair');
  expect(last('/Calendar/events').query.get('startDate')).toBe('2026-10-01');
  expect(last('/Calendar/events').query.get('endDate')).toBe('2026-10-31');
  expect((await success(args)).stdout).toContain('School fair');
  const counts = await success(['notifications', 'counts', '--term', 'term-1']);
  expect(counts.stdout).toContain('9007199254740993');
  expect(JSON.parse(counts.stdout).UnreadMessages).toBe(2);
  expect(last('/Notifications/GetNotificationCounts').query.get('academicTermID')).toBe('term-1');
  w.state.response = 'html';
  await failure(['notifications', 'counts', '--term', 'term-1'], 1, 'not JSON');
});

test('directory flow discovers defaults and retains pagination, search, sort, and academic levels', async () => {
  expect((await json(['directory', 'list']))[0].ID).toBe('directory-1');
  expect((await success(['directory', 'list'])).stdout).toContain('Families');
  const args = ['directory', 'families', '--page', '2', '--per-page', '2', '--search', 'Example & family', '--sort', 'FamilyName', '--ascending', '--academic-level', 'level-1,level-2'];
  expect((await json(args)).FamilyList[0].FamilyId).toBe('family-1');
  const query = last('/Directory/GetFamilyDirectoryUserInfo').query;
  expect(Object.fromEntries(query)).toEqual({ AcademicLevelIds: 'level-2', AscendingOrder: 'true', DirectoryId: 'directory-1', Page: '2', PerPage: '2', SchoolYearId: 'year-1', SearchTerm: 'Example & family', SortProperty: 'FamilyName' });
  expect(query.getAll('AcademicLevelIds')).toEqual(['level-1', 'level-2']);
  const result = await success(args);
  expect(result.stdout).toContain('Pat Example');
  expect(result.stdout).toContain('Page 2 of 4, 8 total');
  const before = w.requests.length;
  await json(['directory', 'families', '--directory', 'explicit-dir', '--school-year', 'explicit-year']);
  expect(w.requests.slice(before).map((r) => r.path)).toEqual(['/Login', '/Login', '/Directory/GetFamilyDirectoryUserInfo']);
  expect(last('/Directory/GetFamilyDirectoryUserInfo').query.get('DirectoryId')).toBe('explicit-dir');
  expect(last('/Directory/GetFamilyDirectoryUserInfo').query.get('SchoolYearId')).toBe('explicit-year');
});

test('message flow forwards paging filters, reads thread content, and persists exact attachment bytes', async () => {
  const args = ['messages', 'list', '--label', 'Archive', '--search', 'field & trip', '--page', '2'];
  expect((await json(args)).MessageThreads[0].MessageThread.ID).toBe('thread-1');
  expect(JSON.parse(last('/Messages/GetMessageThreads').body)).toEqual({ Label: 'Archive', MessageThreadID: '', Page: 2, ProxyUserID: null, SearchTerm: 'field & trip' });
  expect((await success(args)).stdout).toContain('Page 2 of 3, 6 total');
  expect((await json(['messages', 'get', 'thread-1'])).MessageThreadViewModel.Messages).toHaveLength(1);
  expect(JSON.parse(last('/Messages/GetThreadMessages').body)).toEqual({ messageThreadID: 'thread-1', userID: null });
  expect((await success(['messages', 'get', 'thread-1'])).stdout).toContain('Bring & pack lunch.');
  const path = join(w.dir, 'attachment.bin');
  expect(await json(['messages', 'download', 'thread-1', 'file-1', '-o', path])).toEqual({ path, bytes: 6 });
  expect((await readFile(path)).equals(fileBytes)).toBe(true);
  await failure(['messages', 'download', 'thread-1', 'unknown', '-o', join(w.dir, 'missing.bin')], 1, 'not in thread');
  await absent(join(w.dir, 'missing.bin'));
  expect((await success(['messages', 'download', 'thread-1', 'fallback', '-o', path, '--force'])).stdout.trim()).toBe(path);
  expect((await readFile(path)).equals(fileBytes)).toBe(true);
});

test('document flow downloads binary data, refuses overwrite and symlinks, and supports forced fallback URLs', async () => {
  const list = await json(['documents', 'list', '--folder', 'folder-1']);
  expect(list.SchoolDocumentsListItems[0].ID).toBe('document-1');
  expect(last('/SchoolDocuments').query.get('folderID')).toBe('folder-1');
  expect((await success(['documents', 'list'])).stdout).toContain('Handbook');
  const path = join(w.dir, 'document.bin');
  expect(await json(['documents', 'download', 'document-1', '--folder', 'folder-1', '-o', path])).toEqual({ path, bytes: 6 });
  expect((await stat(path)).mode & 0o777).toBe(0o600);
  expect((await readFile(path)).equals(fileBytes)).toBe(true);
  const before = w.requests.filter((r) => r.path === '/SchoolDocuments').length;
  await failure(['documents', 'download', 'document-1', '-o', path], 1, 'output exists');
  expect(w.requests.filter((r) => r.path === '/SchoolDocuments').length).toBe(before);
  expect((await readFile(path)).equals(fileBytes)).toBe(true);
  const link = join(w.dir, 'link.bin');
  await symlink(path, link);
  await failure(['documents', 'download', 'document-1', '-o', link, '--force'], 1, 'non-regular output path');
  expect((await readFile(path)).equals(fileBytes)).toBe(true);
  const unknown = join(w.dir, 'unknown.bin');
  await failure(['documents', 'download', 'unknown', '-o', unknown], 1, 'not in the selected folder');
  await absent(unknown);
  expect((await success(['documents', 'download', 'fallback', '-o', path, '--force'])).stdout.trim()).toBe(path);
  for (const mode of ['error', 'truncated']) {
    w.state.download = mode;
    const fresh = join(w.dir, `failed-${mode}.bin`);
    await failure(['documents', 'download', 'document-1', '-o', fresh], 1, mode === 'error' ? 'download unavailable' : 'read response');
    await absent(fresh);
    await failure(['documents', 'download', 'document-1', '-o', path, '--force'], 1, mode === 'error' ? 'download unavailable' : 'read response');
    expect((await readFile(path)).equals(fileBytes)).toBe(true);
  }
});

test('raw GET merges repeated queries, preserves binary stdout and JSON numbers, and rejects external redirects', async () => {
  const result = await w.run(['--json', '--trace-http', 'raw', 'get', '/raw?tag=existing', '--query', 'tag=next', '--query', 'search=a=b & c']);
  expect(result.code, result.stderr).toBe(0);
  expect(result.stdout).toContain('9007199254740993');
  expect(last('/raw').query.getAll('tag')).toEqual(['existing', 'next']);
  expect(last('/raw').query.get('search')).toBe('a=b & c');
  expect(result.stderr).toContain('[http] GET /raw -> 200');
  expect(result.stderr).not.toContain('synthetic-password');
  expect(result.stderr).not.toContain('authenticated-token');
  expect((await success(['raw', 'get', '/file.bin'])).bytes.equals(fileBytes)).toBe(true);
  await failure(['raw', 'get', '/redirect'], 1, 'refusing redirect outside');
  w.state.response = 'html';
  await failure(['--json', 'raw', 'get', '/raw'], 1, 'not JSON');
});

test('HTTP failures, malformed data, and deadlines return failure without primary output or files', async () => {
  w.state.response = 'http-error';
  await failure(['--json', 'students', 'list'], 1, 'HTTP 503: fixture unavailable');
  const path = join(w.dir, 'failed.bin');
  await failure(['documents', 'download', 'document-1', '-o', path], 1, 'HTTP 503');
  await absent(path);
  w.state.response = 'html';
  await failure(['--json', 'students', 'list'], 1, 'decode response JSON');
  w.state.response = 'timeout';
  await failure(['--timeout', '30ms', 'students', 'list'], 1, 'request failed');
  const unreachable = await w.run(['login'], { env: { CLASSREACH_BASE_URL: 'http://127.0.0.1:1' } });
  expect(unreachable.code).toBe(1);
  expect(unreachable.stdout).toBe('');
  expect(unreachable.stderr).toContain('login request failed');
});

test('invalid invocations fail before authentication, including during dry-run', async () => {
  const cases = [
    ['unknown'], ['--unknown'], ['--json', '--plain', 'version'], ['--timeout', '0s', 'overview'],
    ['overview', 'extra'], ['overview', '--week', 'bad'], ['students', 'get'], ['courses', 'get'],
    ['assignments', 'list'], ['assignments', 'get', 'id', '--student', 'student-1'], ['attendance', 'list'],
    ['messages', 'list', '--page', '0'], ['messages', 'download', 'thread', 'file'],
    ['documents', 'download', 'doc'], ['agenda', 'download', '--output', join(w.dir, 'out'), '--week', 'bad'],
    ['calendar', 'list', '--start', '2026-10-05', '--end', '2026-10-01'],
    ['directory', 'families', '--per-page', '0'], ['notifications', 'counts'], ['notifications', 'counts', '--term', ' '],
    ['raw', 'get', '/raw', '--query', 'bad'], ['completion', 'invalid'],
  ];
  for (const args of cases) {
    await failure(args, 2, 'error:');
    await failure(['--dry-run', ...args], 2, 'error:');
  }
  for (const env of [{ CLASSREACH_TIMEOUT: 'invalid' }, { CLASSREACH_OUTPUT: 'invalid' }, { CLASSREACH_DRY_RUN: 'invalid' }]) {
    expect((await w.run(['version'], { env })).code).toBe(2);
  }
  expect(w.requests).toHaveLength(0);
});

test('dry-run covers every authenticated leaf and filesystem mutation with no network or changes', async () => {
  const output = join(w.dir, 'dry-output');
  const before = await readFile(w.config);
  const cases = [
    ['login'], ['doctor'], ['overview'], ['students', 'list'], ['students', 'get', 'id'], ['courses', 'list'], ['courses', 'get', 'id'],
    ['assignments', 'list', '--student', 'id', '--section', 'id'], ['assignments', 'get', 'id', '--student', 'id', '--section', 'id'],
    ['grades', 'list'], ['attendance', 'list', '--student', 'id', '--section', 'id'], ['messages', 'list'], ['messages', 'get', 'id'],
    ['messages', 'download', 'id', 'id', '-o', output], ['documents', 'list'], ['documents', 'download', 'id', '-o', output],
    ['agenda', 'download', '-o', output], ['announcements', 'list'], ['notifications', 'counts', '--term', 'id'],
    ['calendar', 'list'], ['directory', 'list'], ['directory', 'families'], ['raw', 'get', '/raw'], ['config', 'init', '--force'],
  ];
  for (const args of cases) {
    const result = await json(['--dry-run', ...args]);
    expect(result.dry_run).toBe(true);
    expect(result.command).toContain('classreach');
  }
  const envDry = await w.run(['documents', 'download', 'id', '-o', output], { env: { CLASSREACH_DRY_RUN: 'true', CLASSREACH_OUTPUT: 'json' } });
  expect(envDry.code).toBe(0);
  expect(JSON.parse(envDry.stdout).dry_run).toBe(true);
  expect((await readFile(w.config)).equals(before)).toBe(true);
  await absent(output);
  expect(w.requests).toHaveLength(0);
});

test('version output and shell completion work without credentials and honor environment output overrides', async () => {
  expect((await json(['version'])).version).toBe('dev');
  expect((await success(['--plain', 'version'])).stdout).toBe('dev\n');
  expect((await success(['--version', 'overview'])).stdout).toContain('classreach version dev');
  for (const shell of ['bash', 'zsh', 'fish', 'powershell']) expect((await success(['completion', shell])).stdout).toContain('classreach');
  const envOutput = await w.run(['version'], { env: { CLASSREACH_OUTPUT: 'json' } });
  expect(envOutput.code).toBe(0);
  expect(JSON.parse(envOutput.stdout).version).toBe('dev');
  expect((await w.run(['--plain', 'version'], { env: { CLASSREACH_OUTPUT: 'json' } })).stdout).toBe('dev\n');
  expect(w.requests).toHaveLength(0);
});

// Small immutable ZIP fixtures protect archive delivery and extraction safety.
test('agenda flow preserves ZIP bytes, extracts PDFs, refuses overwrite, and rejects traversal', async () => {
  w.state.agenda = Buffer.from('UEsDBBQAAAAIAKphRF1MJh+IIgAAACAAAAAQAAAAc3R1ZGVudC93ZWVrLnBkZlMNcHHTNdQzUSiuzCvJSC3JTFYoT03NzqlUKM5ITS3hAgBQSwECFAMUAAAACACqYURdTCYfiCIAAAAgAAAAEAAAAAAAAAAAAAAAgAEAAAAAc3R1ZGVudC93ZWVrLnBkZlBLBQYAAAAAAQABAD4AAABQAAAAAAA=', 'base64');
  const zip = join(w.dir, 'agenda.zip');
  expect(await json(['agenda', 'download', '--week', '2026-10-05', '-o', zip])).toEqual([zip]);
  expect((await readFile(zip)).equals(w.state.agenda)).toBe(true);
  const dir = join(w.dir, 'sheets');
  const pdf = join(dir, 'student-week.pdf');
  expect(await json(['agenda', 'download', '-o', dir])).toEqual([pdf]);
  expect(await readFile(pdf, 'utf8')).toBe('%PDF-1.4 synthetic weekly sheet\n');
  expect((await stat(pdf)).mode & 0o777).toBe(0o600);
  await failure(['agenda', 'download', '-o', dir], 1, 'output exists');
  expect((await success(['agenda', 'download', '-o', dir, '--force'])).stdout.trim()).toBe(pdf);
  w.state.agenda = Buffer.from('UEsDBBQAAAAIAKphRF1MJh+IIgAAACAAAAANAAAALi4vZXNjYXBlLnBkZlMNcHHTNdQzUSiuzCvJSC3JTFYoT03NzqlUKM5ITS3hAgBQSwECFAMUAAAACACqYURdTCYfiCIAAAAgAAAADQAAAAAAAAAAAAAAgAEAAAAALi4vZXNjYXBlLnBkZlBLBQYAAAAAAQABADsAAABNAAAAAAA=', 'base64');
  const unsafe = join(w.dir, 'unsafe');
  await failure(['agenda', 'download', '-o', unsafe], 1, 'unsafe ZIP entry');
  await absent(unsafe);
  await absent(join(w.dir, 'escape.pdf'));
  w.state.agenda = Buffer.from('not a ZIP');
  await failure(['agenda', 'download', '-o', join(w.dir, 'bad-zip')], 1, 'read agenda ZIP');
});

test('default config uses isolated home and survives a separate login process', async () => {
  const result = await w.run(['--json', 'config', 'init', '--username', 'synthetic-user', '--password-stdin', '--base-url', w.baseURL, '--origin-host', 'localhost'], { config: null, stdin: ' synthetic-password \n' });
  expect(result.code, result.stderr).toBe(0);
  const path = JSON.parse(result.stdout).path;
  const expected = process.platform === 'darwin' ? join(w.dir, 'Library', 'Application Support', 'classreach', 'config.yaml') : join(w.dir, 'classreach', 'config.yaml');
  expect(path).toBe(expected);
  expect((await stat(path)).mode & 0o777).toBe(0o600);
  const show = await w.run(['--json', 'config', 'show'], { config: null });
  expect(show.code).toBe(0);
  expect(JSON.parse(show.stdout).path).toBe(path);
  const login = await w.run(['--json', 'login'], { config: null });
  expect(login.code, login.stderr).toBe(0);
  expect(JSON.parse(login.stdout).ok).toBe(true);
});
