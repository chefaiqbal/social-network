#!/bin/bash

# Remove duplicate migration files
rm -f pkg/db/migrations/sqlite/00001_users.up.sql
rm -f pkg/db/migrations/sqlite/00001_users.down.sql
rm -f pkg/db/migrations/sqlite/00003_post.up.sql
rm -f pkg/db/migrations/sqlite/00003_post.down.sql

# Remove unnecessary files
rm -f pkg/db/sqlite/db.sql

# Rename migration files to follow consistent naming
mv pkg/db/migrations/sqlite/000001_create_users_table.up.sql pkg/db/migrations/sqlite/000001_users.up.sql
mv pkg/db/migrations/sqlite/000002_create_posts_table.up.sql pkg/db/migrations/sqlite/000002_posts.up.sql

# Make the script executable
chmod +x cleanup.sh

echo "Cleanup completed!" 