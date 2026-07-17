# Releasing bbz

Releases are triggered by pushing a git tag, followed by one manual review step.

1. Pick the next version, following the existing `vX.Y` / `vX.Y.Z` tags (`git tag --sort=-creatordate | head`).
2. Tag and push:

   ```bash
   git tag v0.10
   git push origin v0.10
   ```

3. GitHub Actions (`.github/workflows/release.yaml`) builds binaries for linux/windows/darwin with GoReleaser and creates a **draft** GitHub Release with the archives attached and an auto-generated changelog (from commit/PR titles since the last tag).
4. Review the draft release on GitHub, then click **Publish release**.
5. Publishing fires `.github/workflows/homebrew.yml`, which pushes an updated formula to [ivanizag/homebrew-tap](https://github.com/ivanizag/homebrew-tap), so `brew install ivanizag/tap/bbz` picks it up. Nothing is pushed to Homebrew while the release stays a draft.

That's it — no version bump in code, no changelog file to edit.

## Requirements

- The `TAP_GITHUB_TOKEN` repo secret must be set (a PAT with push access to `ivanizag/homebrew-tap`), or the homebrew.yml step will fail.
