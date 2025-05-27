# Docker Deployment Guide

This guide explains how to set up automated Docker image publishing to Docker Hub using GitHub Actions.

## Setting Up GitHub Secrets

To publish Docker images to Docker Hub, you need to set up the following secrets in your GitHub repository:

1. Go to your repository on GitHub
2. Navigate to Settings > Secrets and variables > Actions
3. Add the following secrets:

| Secret Name | Description |
|-------------|-------------|
| `DOCKER_HUB_USERNAME` | Your Docker Hub username |
| `DOCKER_HUB_TOKEN` | A Docker Hub personal access token (not your password) |

## Creating a Docker Hub Access Token

1. Log in to [Docker Hub](https://hub.docker.com/)
2. Click on your username in the top-right corner
3. Select "Account Settings"
4. Navigate to "Security"
5. Click "New Access Token"
6. Give it a name (e.g., "GitHub Actions")
7. Select "Read & Write" permissions
8. Copy the generated token immediately (it won't be shown again)
9. Add this token as the `DOCKER_HUB_TOKEN` secret in your GitHub repository

## Using the Docker Image

After the GitHub Actions workflow successfully runs, you can use the Docker image as follows:

```bash
# Pull the latest image
docker pull your-username/emailvalidator:latest

# Run the container
docker run -p 8080:8080 your-username/emailvalidator:latest
```

Replace `your-username` with your actual Docker Hub username.

## Versioning

The Docker image is tagged automatically based on:

- Latest commit on main branch: `latest` tag
- Git tags (e.g., v1.0.0): Version-specific tags
- Branches: Branch-specific tags

To create a new versioned release:

1. Create and push a new Git tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. This will trigger the GitHub Actions workflow to build and publish a Docker image with the version-specific tag.

## Troubleshooting

If you encounter issues with the Docker image:

1. Check the GitHub Actions workflow logs for any errors
2. Verify that your Docker Hub credentials are correct
3. Ensure that you have sufficient permissions on Docker Hub
4. Try pulling a specific tag instead of `latest` if you're having caching issues 