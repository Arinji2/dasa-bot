# This script updates all ranks with a specific college ID to a new college ID
#!/bin/bash

OLD_COLLEGE_ID="98j9ibh28z4inw9"
NEW_COLLEGE_ID="1koju2xg7o71as8"

# Step 1: Fetch ranks with OLD_COLLEGE_ID
records=$(curl -s -X GET "http://127.0.0.1:8090/api/collections/ranks/records?filter=college='${OLD_COLLEGE_ID}'&perPage=1000")

# Step 2: Loop through each record and update the college field
echo "$records" | jq -r '.items[].id' | while read -r record_id; do
  echo "Updating record $record_id → $NEW_COLLEGE_ID"
  curl -s -X PATCH "http://127.0.0.1:8090/api/collections/ranks/records/${record_id}" \
    -H "Content-Type: application/json" \
    -d "{\"college\": \"${NEW_COLLEGE_ID}\"}"
done
