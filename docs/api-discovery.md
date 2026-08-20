# ClassReach private API discovery

Status: initial discovery complete

This document contains sanitized endpoint information. It excludes credentials, cookies, tokens,
user IDs, student IDs, section IDs, and response data.

## Hosts

| Purpose | Host |
|---|---|
| Configured tenant | `{school}.classreach.com` |
| Web application origin | `classreach.azurewebsites.net` |
| Mobile notification API | `classreachapi.azurewebsites.net` |

Cloudflare protects the tenant hostname. A normal non-browser request to the tenant login page
returns a managed challenge.

The web application runs on Azure App Service. A request can keep the tenant URL and TLS server
name while the HTTP transport connects to `classreach.azurewebsites.net`. This returns the tenant
application without a Cloudflare challenge. The CLI must apply this routing only to a configured
`classreach.com` host.

## Web authentication

### Login page

```http
GET /Login?ReturnUrl=/
```

The response sets an ASP.NET anti-forgery cookie and contains this form contract:

```text
POST /Login?ReturnUrl=/
Content-Type: application/x-www-form-urlencoded

Username
Password
__RequestVerificationToken
```

A successful login sets `.AspNet.SharedCookie` and redirects away from the login form. Invalid
credentials return the login form with HTTP 200.

Relevant cookie names include:

- `.AspNet.SharedCookie`
- `.AspNetCore.Antiforgery.*`
- `ARRAffinity`
- `ARRAffinitySameSite`
- `currentRole`
- `academicTerm`

The CLI keeps cookies and the post-login anti-forgery token in memory for one process. It does
not log their values. A plain Go client has completed login and guardian reads through the Azure
origin without browser state.

## Mobile API

The Android app identifies this API through its embedded configuration:

```text
https://classreachapi.azurewebsites.net
```

Discovered endpoints:

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/Accounts` | Find accounts and request an authorization token |
| GET | `/api/Roles` | List roles for an authenticated user |
| GET | `/api/Notifications` | List notification summaries |
| POST | `/api/Notifications/MarkAsRead` | Mark notifications as read |
| GET/POST/DELETE | `/api/Devices/*` | Manage push notification devices and settings |

The mobile app puts email and password values in `/api/Accounts` query parameters. The CLI will
use the web login form instead because it sends credentials in a POST body.

The mobile API focuses on notifications. Notification entries link back to the web application.
It does not replace the web endpoints required for messages, documents, grades, and assignments.

## Guardian web endpoints

### Overview

| Method | Path | Parameters |
|---|---|---|
| GET | `/Home/GetQuickView` | `weekDate` |
| GET | `/Home/GetDashboardInfo` | user, role, term, page, read state, skip, take |
| GET | `/Home/GetNotifications` | dashboard filter and paging fields |
| GET | `/Calendar/events` | `startDate`, `endDate` |
| GET | `/Notifications/GetNotificationCounts` | `academicTermID` |
| GET | `/EntitySearch/GetAcademicTermByID` | term lookup fields |

Dashboard read endpoints do not mark notifications as read. Separate POST endpoints perform read
state changes.

### Students and courses

The dashboard response includes visible students and their sections. Course pages use this route:

```text
/Students/{studentID}/Sections/{sectionID}
```

Section landing endpoints:

| Method | Path |
|---|---|
| GET | `/Students/{studentID}/Sections/{sectionID}/SectionLandingPage/GetSectionLinks` |
| GET | `/Students/{studentID}/Sections/{sectionID}/SectionLandingPage/GetSectionItems` |

### Messages

| Method | Path | Notes |
|---|---|---|
| POST | `/Messages/GetMessageThreads` | List and page message threads |
| POST | `/Messages/GetThreadMessages` | Get one thread and its messages |
| POST | `/Messages/GetUsersForMessageRecipient` | Resolve recipient display information |

The web application uses separate POST endpoints for read state, archive state, labels, drafts,
and sends. Version `0.1.0` will not call those mutation endpoints.

`GetThreadMessages` marks an unread thread as read and returns the updated unread count. A live
before-and-after test confirmed this behavior. The CLI and agent skill state this side effect.

### Documents

| Method | Path | Parameters |
|---|---|---|
| GET | `/SchoolDocuments` | `folderID`, page, search term |
| GET | `/SchoolDocumentsFolders/GetSchoolDocumentsFolders` | folder search fields |
| GET | `/SchoolDocumentsFolders/GetSchoolDocumentsFolderByID` | folder ID |

The `/Documents` page provides the injected endpoint URLs and file metadata. Live discovery found
81 visible documents across 19 folders. Each document supplied a `DownloadUrl`; discovery did not
download any file.

### Directory

| Method | Path | Parameters |
|---|---|---|
| GET | `/Directory/GetDirectoryInfo` | none |
| GET | `/Directory/GetDirectoryUserInfo` | directory, school year, search, sort, paging |
| GET | `/Directory/GetFamilyDirectoryUserInfo` | directory, school year, search, sort, paging |
| POST | `/Directory/GetDirectoryCsvUrl` | directory filter model |

The first release will use only the GET endpoints.

### Calendar and agenda

| Method | Path | Parameters |
|---|---|---|
| GET | `/Calendar/events` | `startDate`, `endDate` |
| GET | `/Agenda/GetAgendaForWeek` | week and student context |
| GET | `/Agenda/DownloadAgendaForWeek` | `weekDate` |

The quick view supplies the agenda download URL. The download response is a ZIP archive of PDF
assignment sheets. `agenda download` follows that URL and can save the archive or extract its files.

`/Agenda/ChangeAgendaItemCompletionStatus` is a mutation and remains unsupported.

### Grades, assignments, and attendance

Primary pages:

```text
/Students/{studentID}/Sections/{sectionID}/Grades
/Students/{studentID}/Sections/{sectionID}/Assignments
/Students/{studentID}/Sections/{sectionID}/LessonPlans
/Students/{studentID}/Sections/{sectionID}/Handouts
/Students/{studentID}/Sections/{sectionID}/Attendance
/Students/{studentID}/Sections/{sectionID}/Discussions
```

Observed supporting endpoints include:

```text
GET  /Students/{studentID}/Sections/{sectionID}/GradebookSettings/GetGradebookSettingsInfo
GET  /Students/{studentID}/Sections/{sectionID}/EntitySearch/GetGradingUnitsByString
GET  /Students/{studentID}/Sections/{sectionID}/EntitySearch/GetGradingUnitByID
POST /Students/{studentID}/Sections/{sectionID}/Grades/GetStudentGradeInfoForUnit
GET  /Students/{studentID}/Sections/{sectionID}/Discussions/GetSectionDiscussionInfo
```

Assignments, attendance, lesson plans, and handouts use JSON models embedded in authenticated HTML.
The CLI decodes the assignment and attendance models. Course grade summaries come from the guardian
quick view. The grade unit endpoint returns an HTML fragment; all 24 current-unit responses in the
live account were empty during discovery.

### Other guardian resources

| Method | Path |
|---|---|
| POST | `/Forms/GetFormsSummaryPageInfo` |
| POST | `/SchoolDiscussions/GetSchoolDiscussionsListInfo` |
| GET | `/Financial/GetCustomerPageInfo` |
| GET | `/FinancialAgreements/GetFinancialAgreementsPageInfo` |

These resources remain outside the first typed command slice unless an OpenClaw workflow needs
them.

## Typed CLI coverage

Live verification succeeded for:

- `overview`
- `students list/get`
- `courses list/get`
- `assignments list`
- `grades list`
- `attendance list`
- `messages list/get`
- `documents list`
- `announcements list`
- `calendar list`
- `directory list/families`

## Remaining discovery

- Verify an explicit document download after the user identifies a safe sample.
- Capture non-empty assignment and detailed grade models when the account supplies them.
- Confirm paging fields for lists that exceed the current live result size.
