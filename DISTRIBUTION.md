# Distribution Setup Guide

## Current Status

✅ **Local Development**: Works with ad-hoc signature
❌ **Distribution to Other Users**: Requires Apple Developer account and setup

## For Distribution to Other Users

### Prerequisites

1. **Apple Developer Account** ($99/year)
   - Sign up at https://developer.apple.com

2. **Developer ID Certificate**
   - Log in to Apple Developer portal
   - Go to Certificates, Identifiers & Profiles
   - Create a "Developer ID Application" certificate
   - Download and install in Keychain

### Setting Up GitHub Secrets

For automated signing in GitHub Actions, add these secrets to your repository:

1. **MACOS_CERTIFICATE**
   ```bash
   # Export certificate from Keychain as .p12 file
   # Then convert to base64
   base64 -i YourCertificate.p12 | pbcopy
   # Paste the output into GitHub secret
   ```

2. **MACOS_CERTIFICATE_PWD**
   - The password you set when exporting the .p12 file

3. **APPLE_ID**
   - Your Apple ID email (e.g., hello@tinysystems.io)

4. **APPLE_APP_PASSWORD**
   - Generate app-specific password at https://appleid.apple.com
   - Account Settings → Security → App-Specific Passwords

5. **APPLE_TEAM_ID**
   - Find at https://developer.apple.com/account
   - Membership Details → Team ID

### How to Add GitHub Secrets

1. Go to your repository on GitHub
2. Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Add each secret listed above

## Without Proper Signing

If you distribute the app without proper signing and notarization:
- ❌ Users will get "App is damaged and can't be opened" errors
- ❌ Users will need to manually bypass security (not recommended)
- ❌ Your app won't be trusted by macOS Gatekeeper

## With Proper Signing & Notarization

- ✅ Users can install and run without warnings
- ✅ App is trusted by macOS Gatekeeper
- ✅ Professional distribution

## Potential PATH Issues

The app checks these locations for `gke-gcloud-auth-plugin`:
- `~/google-cloud-sdk/bin` (most common)
- `~/.local/bin`
- `~/go/bin`
- `~/.krew/bin`
- `/usr/local/bin`
- `/opt/homebrew/bin`

**Limitation**: If users install gcloud in non-standard locations, they may still encounter issues.

**Alternative Solution**: Bundle `gke-gcloud-auth-plugin` with your app (increases app size but guarantees availability).

## Testing Before Release

Before distributing to users:

1. Test the signed DMG on a different Mac
2. Verify the app can find `gke-gcloud-auth-plugin`
3. Check that GKE authentication works
4. Test with users who have gcloud installed via Homebrew vs direct download

## Current Workflow Behavior

The updated GitHub Actions workflow will:
- ✅ Use entitlements for all builds
- ✅ Use proper Developer ID signature if secrets are set
- ✅ Notarize the DMG if credentials are provided
- ⚠️  Fall back to ad-hoc signature if secrets are not set (development only)

## Cost Summary

- **Apple Developer Program**: $99/year (required for distribution)
- **Everything else**: Free
