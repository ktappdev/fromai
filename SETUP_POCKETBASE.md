# PocketBase Setup Instructions

## Create Tasks Collection

1. Open http://127.0.0.1:8090/_/ in your browser
2. Login with:
   - Email: admin@example.com
   - Password: password123
3. Go to Settings > Collections
4. Click "New Collection"
5. Name: `tasks`
6. Type: Base
7. Add the following fields:

### Fields:
- **title** (text, required)
- **description** (text, not required)
- **starter_code** (text, required)
- **code** (text, required)
- **language** (text, required, max: 50, pattern: `^(typescript|javascript|python|go|rust|java|cpp)$`)
- **status** (select, required, values: "pending", "completed")
- **grade** (text, not required)
- **feedback** (text, not required)
- **user** (relation, required, collection: users, max select: 1)
- **created_at** (number, required)
- **updated_at** (number, required)

### Indexes:
- `by_user` on field `user`
- `by_user_status` on fields `user` and `status`

8. Save the collection

## API Rules
Set all API rules to allow access for authenticated users (or leave empty for public access during testing).

## Start PocketBase
```bash
./pocketbase serve
```

## Start SvelteKit
```bash
pnpm install
pnpm dev
```

The app will be available at http://localhost:5173
