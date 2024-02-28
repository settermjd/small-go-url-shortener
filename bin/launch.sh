#!/bin/sh

# See https://www.gnu.org/software/bash/manual/html_node/The-Set-Builtin.html for more information
set -Cu

# Run the database migrations
/opt/bin/run-migrations.sh

# Launch the app in the foreground
gourlshortener