#!/bin/bash
# go back
cd ../

# remove the file
rm -rf 5003

# Define the file path
file_path="constants/constants.go"

# Use sed to search and replace the value
sed -i 's/\(BLOCKCHAIN_DB_PATH\s*=\s*"\)[^\/]*\/evodb"/\15003\/evodb"/' "$file_path"

# run the file
go run main.go chain -port 5003 -miners_address evochain42d40be8b315e31dac50a4daf93ce366b1c62668 -remote_node http://127.0.0.1:5000