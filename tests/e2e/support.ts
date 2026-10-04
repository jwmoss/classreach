import { spawn } from 'node:child_process';
import { createServer } from 'node:http';
import { mkdtemp, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve, join } from 'node:path';

const loginForm = '<input name="__RequestVerificationToken" value="login-token"><input name="Username"><input name="Password">';
const student = { UserID: 'student-1', Name: 'Alex Example', Sections: [{ Course: { ID: 'course-1', Name: 'Science' }, Section: { ID: 'section-1', AcademicTerm_ID: 'term-1' }, Grade: 94.5, LetterGrade: 'A', SectionUrl: '/section' }] };
const assignment = { Assignment: { ID: 'assignment-1', Name: "Earth's orbit", Description: '<p>Explain &amp; compare.</p>' }, AssignmentState: 'Assigned', SectionGradingCategoryName: 'Homework' };
export const fileBytes = Buffer.from([0, 255, 80, 68, 70, 10]);

export async function workspace() {
  const dir = await mkdtemp(join(tmpdir(), 'classreach-e2e-'));
  const config = join(dir, 'config.yaml');
  const requests: any[] = [];
  const state = { login: 'ok', response: 'ok', download: 'ok', agenda: Buffer.alloc(0) };
  const server = createServer(async (req, res) => {
    const url = new URL(req.url!, 'http://localhost');
    const chunks = [];
    for await (const chunk of req) chunks.push(chunk);
    const body = Buffer.concat(chunks).toString();
    requests.push({ method: req.method, path: url.pathname, query: url.searchParams, body, headers: req.headers });
    const json = (value: unknown) => { res.setHeader('Content-Type', 'application/json'); res.end(JSON.stringify(value)); };
    const htmlModel = (marker: string, value: unknown) => res.end(`<script>${marker} = JSON.parse('${JSON.stringify(JSON.stringify(value)).slice(1, -1).replaceAll("'", "\\'")}');</script>`);
    if (url.pathname === '/Login') {
      if (req.method === 'GET') return res.end(state.login === 'missing-token' ? '<html>Challenge</html>' : loginForm);
      const form = new URLSearchParams(body);
      if (form.get('Username') !== 'synthetic-user' || form.get('Password') !== ' synthetic-password ' || form.get('__RequestVerificationToken') !== 'login-token') {
        res.statusCode = 400; return res.end('bad login form');
      }
      if (state.login === 'denied') return res.end(loginForm);
      if (state.login !== 'no-session') res.setHeader('Set-Cookie', '.AspNet.SharedCookie=synthetic-session; Path=/; HttpOnly');
      return res.end('<input name="__RequestVerificationToken" value="authenticated-token">');
    }
    if (!req.headers.cookie?.includes('.AspNet.SharedCookie=synthetic-session')) {
      res.statusCode = 401; return json({ message: 'session required' });
    }
    if (req.method === 'POST' && req.headers.__requestverificationtoken !== 'authenticated-token') {
      res.statusCode = 403; return json({ message: 'anti-forgery token required' });
    }
    if (state.response === 'http-error') { res.statusCode = 503; return json({ message: 'fixture unavailable' }); }
    if (state.response === 'html') return res.end('<html>Maintenance</html>');
    if (state.response === 'empty-view') return json({});
    if (state.response === 'timeout') return;
    switch (url.pathname) {
      case '/Home/GetQuickView': return json({ UserInfos: [student, { UserID: 'student-2', Name: 'Casey Example', Sections: [] }], Announcements: [{ Heading: 'School fair', Description: '<b>Bring</b> books &amp; games.', Important: true }], DownloadAgendaForWeekUrl: '/agenda.zip' });
      case '/Students/student-1/Sections/section-1/Assignments': return htmlModel('window.Assignments.model', { AssignmentsList: [assignment] });
      case '/Students/student-1/Sections/section-1/Attendance': return htmlModel('window.AttendanceModule.model', { attendanceMarkings: [{ ID: 'present', Value: 'Present' }], studentAttendance: { Attendance: [{ ID: 'attendance-1', Date: '2026-10-01', AttendanceMarking_ID: 'present' }] } });
      case '/Calendar/events': return json({ CalendarEvents: [{ CalendarEvent: { ID: 'event-1', Name: 'School fair', StartTime: '2026-10-05T12:00:00', EndTime: '2026-10-05T13:00:00' } }] });
      case '/Notifications/GetNotificationCounts': return res.end('{"UnreadMessages":2,"ProviderID":9007199254740993}');
      case '/Directory/GetDirectoryInfo': return json({ Directories: [{ ID: 'directory-1', Name: 'Families', IsFamilyDirectory: true }] });
      case '/Directory': return res.end("<script>window.Directory.SchoolYearForTodayID = 'year-1';</script>");
      case '/Directory/GetFamilyDirectoryUserInfo': return json({ FamilyList: [{ FamilyId: 'family-1', FamilyName: 'Example', GuardianDetails: [{ FullName: 'Pat Example' }], StudentDetails: [{ FullName: 'Alex Example' }] }], PagingInfo: { CurrentPage: Number(url.searchParams.get('Page')), TotalPages: 4, TotalItems: 8 } });
      case '/Messages/GetMessageThreads': return json({ MessageThreads: [{ MessageThread: { ID: 'thread-1', Subject: 'Field trip' }, MessageThreadUserAttributes: { IsRead: false }, TopMessage: { Sender: { FullName: 'Teacher Example' } } }], PagingInfo: { CurrentPage: JSON.parse(body).Page, TotalPages: 3, TotalItems: 6 } });
      case '/Messages/GetThreadMessages': return json({ MessageThreadViewModel: { MessageThread: { ID: 'thread-1', Subject: 'Field trip' }, Messages: [{ Sender: { FullName: 'Teacher Example' }, Message: { Body: '<p>Bring &amp; pack lunch.</p>' }, Files: [{ ID: 'file-1', Url: '/file.bin' }, { ID: 'fallback', FileDownloadLink: '/file.bin' }] }] } });
      case '/SchoolDocuments': return json({ SchoolDocumentsFoldersListItems: [{ ID: 'folder-1', Name: 'Handbooks' }], SchoolDocumentsListItems: [{ ID: 'document-1', Name: 'Handbook', FileInfo: { DownloadUrl: '/file.bin', File: { Size: fileBytes.length } } }, { ID: 'fallback', FileInfo: { File: { FileDownloadLink: '/file.bin' } } }] });
      case '/file.bin':
        if (state.download === 'error') { res.statusCode = 503; return json({ message: 'download unavailable' }); }
        if (state.download === 'truncated') { res.setHeader('Content-Length', 100); res.write(fileBytes); setImmediate(() => res.destroy()); return; }
        return res.end(fileBytes);
      case '/agenda.zip': return res.end(state.agenda);
      case '/raw': return res.end('{"id":9007199254740993,"ok":true}');
      case '/redirect': res.writeHead(302, { Location: 'http://localhost:1/outside' }); return res.end();
      default: res.statusCode = 404; return json({ message: 'fixture route missing' });
    }
  });
  await new Promise<void>((done) => server.listen(0, '127.0.0.1', done));
  const baseURL = `http://127.0.0.1:${(server.address() as any).port}`;
  await writeFile(config, `base_url: ${baseURL}\norigin_host: localhost\nusername: synthetic-user\npassword: ' synthetic-password '\n`, { mode: 0o600 });
  const env = { PATH: process.env.PATH, HOME: dir, XDG_CONFIG_HOME: dir, SystemRoot: process.env.SystemRoot, NO_COLOR: '1', DO_NOT_TRACK: '1' };
  async function run(args: string[], options: { stdin?: string; env?: Record<string, string>; config?: string | null } = {}) {
    return new Promise<{ code: number | null; stdout: string; stderr: string; bytes: Buffer }>((done, reject) => {
      const child = spawn(resolve('bin/classreach'), [...(options.config === null ? [] : ['--config', options.config ?? config]), '--no-input', ...args], { env: { ...env, ...options.env }, stdio: 'pipe' });
      const stdout: Buffer[] = [], stderr: Buffer[] = [];
      const timer = setTimeout(() => { child.kill('SIGKILL'); reject(new Error(`CLI exceeded 5 seconds: ${args.join(' ')}`)); }, 5000);
      child.stdout.on('data', (b) => stdout.push(b));
      child.stderr.on('data', (b) => stderr.push(b));
      child.on('error', (e) => { clearTimeout(timer); reject(e); });
      child.on('close', (code) => { clearTimeout(timer); const bytes = Buffer.concat(stdout); done({ code, stdout: bytes.toString(), bytes, stderr: Buffer.concat(stderr).toString() }); });
      child.stdin.end(options.stdin ?? '');
    });
  }
  return { dir, config, baseURL, state, requests, run, async close() { server.closeAllConnections(); await new Promise<void>((done) => server.close(() => done())); await rm(dir, { recursive: true, force: true }); } };
}
