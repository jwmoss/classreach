# ClassReach capabilities

This CLI supports selected guardian reads from the ClassReach private web API.
The repository's [discovery record](api-discovery.md) supplies the known endpoint paths.
The public support site describes product features, not a complete REST specification.
No current claim of full guardian, teacher, admin, or mobile API coverage is justified.

## Current commands

`S` means `/Students/{studentID}/Sections/{sectionID}`.

| Area | Request | Command | Limits |
| --- | --- | --- | --- |
| Login | GET/POST `/Login?ReturnUrl=/` | Automatic; `login` | Requires the known session cookie; redirects stay on the same origin |
| Overview | GET `/Home/GetQuickView` | `overview` | Guardian quick view |
| Students/courses/grades/announcements | GET `/Home/GetQuickView` | Existing list/get commands | Projections from the quick view; grades are course summaries |
| Assignments | GET `S/Assignments` | `assignments list/get` | Embedded summary model; no submission or attachment contract |
| Attendance | GET `S/Attendance` | `attendance list` | Embedded section model |
| Calendar | GET `/Calendar/events` | `calendar list` | Explicit date range; no writes |
| Agenda files | GET download URL from quick view | `agenda download` | ZIP/PDF assignment sheets, not structured agenda tasks |
| Message list | POST `/Messages/GetMessageThreads` | `messages list` | Manual page, label, and search flags |
| Message detail | POST `/Messages/GetThreadMessages` | `messages get` | Marks an unread thread as read |
| Message file | GET returned file URL | `messages download` | Thread lookup also changes read state |
| Documents | GET `/SchoolDocuments` | `documents list/download` | Selected folder only; page/search contract remains incomplete |
| Directory types | GET `/Directory/GetDirectoryInfo` | `directory list` | Lists directory definitions |
| Family directory | GET `/Directory/GetFamilyDirectoryUserInfo` | `directory families` | Manual page; no automatic complete export |
| Notification counts | GET `/Notifications/GetNotificationCounts?academicTermID=...` | `notifications counts --term ID` | Lossless provider JSON; count field schema remains unverified |
| Known GET request | GET supplied path | `raw get` | No inferred schema or guarantee that a private GET is side-effect-free |

The notification-count request has local boundary tests. It has no new live tenant verification.
Find academic term IDs in the section data from `overview --json`.
An empty or inaccessible result does not prove that another role has no data.

## Known gaps

| Area | Known request or page | Required evidence |
| --- | --- | --- |
| Structured agenda | GET `/Agenda/GetAgendaForWeek` | Exact week/student parameters and nonempty response model |
| Notifications | GET `/Home/GetNotifications`, `/Home/GetDashboardInfo` | Filter fields, role/term context, pagination, and response models |
| Handouts/lesson plans | `S/Handouts`, `S/LessonPlans` | Embedded model markers, detail URLs, and file contracts |
| Detailed grades | POST `S/Grades/GetStudentGradeInfoForUnit` | Nonempty HTML model, unit parameters, and authorization |
| Forms | POST `/Forms/GetFormsSummaryPageInfo` | Request model, page behavior, and response states |
| School discussions | POST `/SchoolDiscussions/GetSchoolDiscussionsListInfo` | Request model and page/detail contracts |
| Section discussions | GET `S/Discussions/GetSectionDiscussionInfo` | Response and thread/post contracts |
| Other directory people | GET `/Directory/GetDirectoryUserInfo` | Exact filters, response models, and page completeness |
| Document folders | GET `/SchoolDocumentsFolders/GetSchoolDocumentsFolders` and `/SchoolDocumentsFolders/GetSchoolDocumentsFolderByID` | Search/lookup parameters and response contracts |
| Finance | GET `/Financial/GetCustomerPageInfo`, `/FinancialAgreements/GetFinancialAgreementsPageInfo` | Account context, response models, and role restrictions |
| Reports/registration/profile | Public product features | Endpoint paths, payloads, and side effects |
| Mobile | Separate `classreachapi.azurewebsites.net` API | Separate authentication and complete request contracts |
| Teacher/admin | Separate role-specific product features | Authorized role access and independent discovery |

Use authorized, redacted samples to close these gaps. Do not commit raw HAR files or school records.
Synthetic fixtures test parsing and request construction; they do not prove a private provider contract.
Do not add arbitrary POST, send, submission, or payment commands from guessed endpoint names.
Verify pagination with multiple pages before adding `--all` or reporting complete results.

## Native calendar subscriptions

ClassReach already supplies a role-selectable ICS share URL.
Open the calendar's Share control, select the roles, and copy the subscription URL.
Subscribe through the destination calendar's native URL subscription feature.
Test event updates, deletions, timezone handling, and refresh delay before relying on the connection.

Treat the subscription URL as a secret. ClassReach can regenerate it to invalidate an old link.
This is a one-way calendar feed. It does not replace agenda assignment-sheet downloads or form reminders.
Do not build a separate synchronization service until the native feed fails a concrete requirement.

Source: [ClassReach calendar sharing](https://help.classreach.com/syncing-your-classreach-calendar-with-external-calendars-like-gmail-or-ical).
See also [guardian features](https://help.classreach.com/guardian-student-documentation),
[teacher features](https://help.classreach.com/teacher-documentation), and
[admin features](https://help.classreach.com/admin-documentation).

## Output and safety contracts

- Each resource command authenticates once. An ordinary read needs no separate login or doctor call.
- `--dry-run` previews resource commands and config creation without network requests or file changes.
- Local inspection commands, such as `version` and `config show`, still return normal output.
- `--json` emits one JSON value or fails. Raw mode preserves bytes when JSON mode is absent.
- Unix file writes use private temporary files and atomic publication. Destination symlinks are rejected.
- Windows inherits directory access permissions. Use a private user directory for configs and downloads.
- HTTP responses have a 64 MiB limit. ZIP extraction permits 1,000 entries and 128 MiB in total.
- Login and API redirects stay on the original scheme and host, including port, with a ten-hop limit.
- A provider file redirect to another origin fails closed. Add a documented download policy only after verifying a real file contract.
