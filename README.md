# ProjectFlow

A project management REST API built with Go.

## Tech Stack

- **Language:** Go
- **Router:** chi
- **Database:** PostgreSQL
- **ORM:** sqlc
- **Migrations:** Goose
- **Auth:** JWT + refresh tokens

## Prerequisites

- Go 1.21+
- PostgreSQL
- [swag](https://github.com/swaggo/swag) for docs generation

## Getting Started

**1. Clone the repository:**

```bash
git clone https://github.com/knnedy/projectflow
cd projectflow
```

**2. Set up environment variables:**

```bash
cp .env.example .env
```

Update `.env` with your values:

```
DB_URL=postgresql://postgres:password@localhost:5432/projectflow
JWT_SECRET=your-secret-key
PORT=8080
ENV=development
```

**3. Run migrations:**

```bash
make migrate-up
```

**4. Run the server:**

```bash
make run
```

The API will be available at `http://localhost:8080`.

Swagger docs are available at `http://localhost:8080/docs`.

## Makefile Commands

| Command                           | Description                  |
| --------------------------------- | ---------------------------- |
| `make run`                        | Build and run the server     |
| `make build`                      | Build the binary             |
| `make migrate-up`                 | Run all pending migrations   |
| `make migrate-down`               | Roll back the last migration |
| `make migrate-reset`              | Roll back all migrations     |
| `make migrate-status`             | Show migration status        |
| `make migrate-create name=<name>` | Create a new migration       |
| `make swagger`                    | Generate swagger docs        |
| `make sqlc`                       | Regenerate sqlc code         |

## API Overview

### Auth

| Method | Endpoint                | Description          |
| ------ | ----------------------- | -------------------- |
| POST   | `/api/v1/auth/register` | Register a new user  |
| POST   | `/api/v1/auth/login`    | Login                |
| POST   | `/api/v1/auth/refresh`  | Refresh access token |
| POST   | `/api/v1/auth/logout`   | Logout               |

### Users

| Method | Endpoint                    | Description            |
| ------ | --------------------------- | ---------------------- |
| GET    | `/api/v1/users/me`          | Get authenticated user |
| PATCH  | `/api/v1/users/me`          | Update profile         |
| PATCH  | `/api/v1/users/me/password` | Update password        |
| DELETE | `/api/v1/users/me`          | Delete account         |

### Organisations

| Method | Endpoint                        | Description         |
| ------ | ------------------------------- | ------------------- |
| POST   | `/api/v1/organisations`         | Create organisation |
| GET    | `/api/v1/organisations`         | List organisations  |
| GET    | `/api/v1/organisations/{orgID}` | Get organisation    |
| PATCH  | `/api/v1/organisations/{orgID}` | Update organisation |
| DELETE | `/api/v1/organisations/{orgID}` | Delete organisation |

### Members

| Method | Endpoint                                            | Description        |
| ------ | --------------------------------------------------- | ------------------ |
| GET    | `/api/v1/organisations/{orgID}/members`             | List members       |
| PATCH  | `/api/v1/organisations/{orgID}/members/{memberID}`  | Update member role |
| DELETE | `/api/v1/organisations/{orgID}/members/{memberID}`  | Remove member      |
| DELETE | `/api/v1/organisations/{orgID}/members/me`          | Leave organisation |
| POST   | `/api/v1/organisations/{orgID}/members/invitations` | Invite member      |
| POST   | `/api/v1/invitations/accept`                        | Accept invitation  |

### Projects

| Method | Endpoint                                             | Description    |
| ------ | ---------------------------------------------------- | -------------- |
| POST   | `/api/v1/organisations/{orgID}/projects`             | Create project |
| GET    | `/api/v1/organisations/{orgID}/projects`             | List projects  |
| GET    | `/api/v1/organisations/{orgID}/projects/{projectID}` | Get project    |
| PATCH  | `/api/v1/organisations/{orgID}/projects/{projectID}` | Update project |
| DELETE | `/api/v1/organisations/{orgID}/projects/{projectID}` | Delete project |

### Issues

| Method | Endpoint                                                                     | Description   |
| ------ | ---------------------------------------------------------------------------- | ------------- |
| POST   | `/api/v1/organisations/{orgID}/projects/{projectID}/issues`                  | Create issue  |
| GET    | `/api/v1/organisations/{orgID}/projects/{projectID}/issues`                  | List issues   |
| GET    | `/api/v1/organisations/{orgID}/projects/{projectID}/issues/{issueID}`        | Get issue     |
| PATCH  | `/api/v1/organisations/{orgID}/projects/{projectID}/issues/{issueID}`        | Update issue  |
| PATCH  | `/api/v1/organisations/{orgID}/projects/{projectID}/issues/{issueID}/status` | Update status |
| DELETE | `/api/v1/organisations/{orgID}/projects/{projectID}/issues/{issueID}`        | Delete issue  |

### Comments

| Method | Endpoint                                                                                   | Description    |
| ------ | ------------------------------------------------------------------------------------------ | -------------- |
| POST   | `/api/v1/organisations/{orgID}/projects/{projectID}/issues/{issueID}/comments`             | Create comment |
| GET    | `/api/v1/organisations/{orgID}/projects/{projectID}/issues/{issueID}/comments`             | List comments  |
| PATCH  | `/api/v1/organisations/{orgID}/projects/{projectID}/issues/{issueID}/comments/{commentID}` | Update comment |
| DELETE | `/api/v1/organisations/{orgID}/projects/{projectID}/issues/{issueID}/comments/{commentID}` | Delete comment |

### Activity

| Method | Endpoint                                                      | Description           |
| ------ | ------------------------------------------------------------- | --------------------- |
| GET    | `/api/v1/organisations/{orgID}/activity`                      | Org activity feed     |
| GET    | `/api/v1/organisations/{orgID}/projects/{projectID}/activity` | Project activity feed |
| GET    | `/api/v1/activity/{entityID}`                                 | Entity activity feed  |

## Authentication

All protected endpoints require a Bearer token in the Authorization header:

```
Authorization: Bearer <access_token>
```

Access tokens expire after 15 minutes. Use the refresh endpoint to get a new one. The refresh token is stored in an httponly cookie and is automatically sent with refresh requests.

## Roles

| Role     | Permissions                                         |
| -------- | --------------------------------------------------- |
| `OWNER`  | Full access including deleting the organisation     |
| `ADMIN`  | Create and manage projects, issues and members      |
| `MEMBER` | Read access, create and update issues, add comments |

## Issue Status Flow

```
BACKLOG → TODO → IN_PROGRESS → IN_REVIEW → DONE
                                          → CANCELLED
```

Status transitions are strictly enforced. Only owners and admins can move issues to `TODO`, `DONE` or `CANCELLED`.
