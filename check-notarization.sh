#!/bin/bash
# Poll Apple notarization status until one succeeds
# Usage: ./check-notarization.sh

echo "Checking notarization status..."
while true; do
  status=$(xcrun notarytool history --keychain-profile "notary" 2>&1)
  accepted=$(echo "$status" | grep -B3 "status: Accepted" | head -4)

  if [ -n "$accepted" ]; then
    echo ""
    echo "NOTARIZATION COMPLETE!"
    echo "$accepted"
    id=$(echo "$accepted" | grep "id:" | awk '{print $2}')
    echo ""
    echo "Run this to staple and re-upload:"
    echo "  xcrun stapler staple tinysystems-darwin-universal.dmg"
    break
  fi

  in_progress=$(echo "$status" | grep -c "In Progress")
  invalid=$(echo "$status" | grep -c "Invalid")
  echo "[$(date +%H:%M:%S)] $in_progress still in progress, $invalid invalid"

  if [ "$invalid" -gt 0 ]; then
    echo ""
    echo "Some submissions were REJECTED. Check logs:"
    echo "$status" | grep -B3 "Invalid"
    echo ""
    echo "Run: xcrun notarytool log <id> --keychain-profile notary"
  fi

  sleep 300
done
