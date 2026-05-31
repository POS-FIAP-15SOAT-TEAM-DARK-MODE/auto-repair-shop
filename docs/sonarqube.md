# SonarQube — how to run

This document explains how to spin up the local SonarQube instance, create the project, and run the analysis.

Summary
- SonarQube runs as a Docker service defined in `docker-compose.yml` and is exposed at `http://localhost:9000`.
- Analysis is triggered via `make sonarqube-run`, which generates a coverage report first and then invokes `sonar-scanner`.
- A project-scoped token stored in `.env` as `SONARQUBE_PROJECT_TOKEN` authenticates the scanner.

Prerequisites
- Docker running and `sonar-scanner` CLI installed locally.
- A `.env` file at the project root (copy from `.env.example` if it does not exist).

Setup (first time only)

1. Start the SonarQube container:

   ```bash
   make docker-up
   ```

2. Open `http://localhost:9000` and log in with `admin / admin`. You will be prompted to set a new password — do so before continuing.

3. Create the project:
   - Click **Projects → Create a local project**.
   - Set both **Project display name** and **Project key** to `auto-repair-shop`.
   - Keep **Main branch name** as `main` and click **Next**.
   - Select **Use the global setting** and click **Create project**.

4. Generate an analysis token:
   - On the project page, choose **Locally** under *How do you want to analyze your repository?*
   - Click **Generate** (set *Expires in* to **No expiration** if preferred).
   - Copy the `sqp_xxxxx` token.

5. Paste the token into `.env`:

   ```env
   SONARQUBE_PROJECT_TOKEN=sqp_xxxxx
   ```

Running the analysis

```bash
make sonarqube-run
```

This runs `test-coverage` first (producing `coverage.out`) and then invokes `sonar-scanner`. Results are available on the project dashboard at `http://localhost:9000`.

Notes
- The `sonar-project.properties` file at the project root configures exclusions (test files, mocks, migrations) and points the scanner at `coverage.out` for Go coverage data.
- Re-running `make sonarqube-run` after code changes will update the dashboard automatically — no need to recreate the project or token.
