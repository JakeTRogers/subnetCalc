FROM ghcr.io/astral-sh/uv:0.13.10@sha256:a3ea1357342c1a4bf9a0c53a3deb1b55379aefa037338d8113d5771efa01774f AS uv

FROM mcr.microsoft.com/devcontainers/go:1.27@sha256:74f45fc4c2abe0de9a9acdb4225ffb3e58e6b28f9af2e209afaa8920a3d3b623

USER root

RUN apt-get update \
    && apt-get install -y --no-install-recommends bc ca-certificates curl gh \
    && case "$(dpkg --print-architecture)" in \
        amd64) arch=x64 ;; \
        arm64) arch=arm64 ;; \
        *) exit 1 ;; \
    esac \
    && curl -fL "https://github.com/github/copilot-cli/releases/latest/download/copilot-linux-${arch}.tar.gz" -o /tmp/copilot.tar.gz \
    && tar -xzf /tmp/copilot.tar.gz -C /usr/local/bin copilot \
    && rm -f /tmp/copilot.tar.gz \
    && rm -rf /var/lib/apt/lists/*

COPY --from=uv /uv /uvx /usr/local/bin/

# Install outside /home/vscode so the tool survives the home volume mount.
RUN export UV_TOOL_DIR=/opt/uv/tools UV_TOOL_BIN_DIR=/usr/local/bin \
    && uv tool install commitizen \
    && uv tool install pre-commit

USER vscode
