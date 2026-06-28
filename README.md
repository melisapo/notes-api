# notes-api

A simple API for sharing positive and uplifting notes. Anyone can post, like, and report notes. A moderator reviews flagged content.

## Endpoints

### Public
| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/posts` | List approved posts (paginated) |
| `GET` | `/posts/{id}` | Get a single post |
| `GET` | `/posts/random` | Get a random post |
| `POST` | `/posts` | Submit a new post |
| `POST` | `/posts/{id}/like` | Like a post (once per IP) |
| `POST` | `/posts/{id}/report` | Report a post |

### Admin
Requires `X-Admin-Key` header.

| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/admin/posts` | List posts by status |
| `PATCH` | `/admin/posts/{id}` | Approve or reject a post |
| `GET` | `/admin/reports` | List reports |
| `DELETE` | `/admin/posts/{id}` | Delete a post |

## Docs
Swagger UI available at https://notes-api-q7ki.onrender.com/swagger/index.html.
