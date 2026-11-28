FROM ghcr.io/astral-sh/uv:0.13.10@sha256:a3ea1357342c1a4bf9a0c53a3deb1b55379aefa037338d8113d5771efa01774f AS uv

FROM mcr.microsoft.com/devcontainers/go:2-1.25

USER root

RUN apt-get update \
    && apt-get install -y --no-install-recommends bc \
    && rm -rf /var/lib/apt/lists/*

COPY --from=uv /uv /uvx /usr/local/bin/

# Install outside /home/vscode so the tool survives the home volume mount.
RUN export UV_TOOL_DIR=/opt/uv/tools UV_TOOL_BIN_DIR=/usr/local/bin \
    && uv tool install commitizen \
    && uv tool install pre-commit

USER vscode
