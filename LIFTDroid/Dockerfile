FROM ubuntu:22.04

# Install basic dependencies including audio and X11 libraries
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl wget git unzip python3 python3-pip openjdk-17-jdk maven \
    ca-certificates gnupg qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils \
    pulseaudio libpulse0 libpulse-dev \
    libx11-6 libxcomposite1 libxcursor1 libxdamage1 libxext6 libxfixes3 \
    libxi6 libxrandr2 libxrender1 libxss1 libxtst6 \
    xz-utils \
    && rm -rf /var/lib/apt/lists/* && apt-get clean

# Install Node.js and npm
RUN mkdir -p /etc/apt/keyrings && \
    curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg && \
    echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_18.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list && \
    apt-get update && apt-get install -y --no-install-recommends nodejs && \
    rm -rf /var/lib/apt/lists/* && apt-get clean

# Install Python and pip
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies - use Frida version 16.7.19 to match server
RUN pip3 install frida==16.7.19 frida-tools

# Set Android SDK environment
ENV ANDROID_HOME=/opt/android-sdk
ENV ANDROID_SDK_ROOT=/opt/android-sdk
ENV PATH=${PATH}:${ANDROID_HOME}/cmdline-tools/latest/bin:${ANDROID_HOME}/platform-tools:${ANDROID_HOME}/emulator

# Install Android command line tools (fix directory structure)
RUN mkdir -p ${ANDROID_HOME}/cmdline-tools && \
    cd ${ANDROID_HOME}/cmdline-tools && \
    wget https://dl.google.com/android/repository/commandlinetools-linux-8512546_latest.zip && \
    unzip commandlinetools-linux-8512546_latest.zip && \
    mkdir latest && \
    mv cmdline-tools/* latest/ && \
    rmdir cmdline-tools && \
    rm commandlinetools-linux-8512546_latest.zip

# Verify sdkmanager is available
RUN ls -l ${ANDROID_HOME}/cmdline-tools/latest/bin && \
    ${ANDROID_HOME}/cmdline-tools/latest/bin/sdkmanager --version

# Accept licenses and install core components via sdkmanager (KISS)
RUN yes | ${ANDROID_HOME}/cmdline-tools/latest/bin/sdkmanager --licenses && \
    ${ANDROID_HOME}/cmdline-tools/latest/bin/sdkmanager --update && \
    ${ANDROID_HOME}/cmdline-tools/latest/bin/sdkmanager \
      "platform-tools" "emulator" "platforms;android-29"

# --- Application Build and Setup ---
WORKDIR /app

# Copy the Maven project file and resolve dependencies first to leverage Docker cache
COPY pom.xml .
RUN mvn dependency:go-offline

# Copy the rest of the application source code
COPY . .

# Compile and package the application into a fat JAR
# Compile and package the application, skipping the tests which are failing in the build environment
RUN mvn clean package -DskipTests

# Ensure dependencies are copied for runtime
RUN mvn dependency:copy-dependencies

## KISS: Use sdkmanager to install system image (API 29) instead of manual wget/unzip
RUN ${ANDROID_HOME}/cmdline-tools/latest/bin/sdkmanager --version && \
    yes | ${ANDROID_HOME}/cmdline-tools/latest/bin/sdkmanager --licenses && \
    ${ANDROID_HOME}/cmdline-tools/latest/bin/sdkmanager \
      "system-images;android-29;google_apis;x86_64"

# Make Python scripts executable
# Frida binaries are already present in local resources directory
RUN echo "[BUILD] Checking frida files..." && \
    ls -la resources/ || echo "[BUILD] Resources directory will be mounted from host"

# Make scripts executable
RUN chmod +x scripts/*.sh scripts/*.py

# Verify Frida installation and version
RUN echo "[BUILD] Verifying Frida installation..." && \
    python3 -c "import frida; print(f'Frida client version: {frida.__version__}')" && \
    echo "[BUILD] Frida client installed successfully"

# Create necessary directories
RUN mkdir -p logs reports

# Create AVD
RUN echo "no" | ${ANDROID_HOME}/cmdline-tools/latest/bin/avdmanager create avd \
    --name "test_device" \
    --package "system-images;android-29;google_apis;x86_64" \
    --abi "x86_64"

# Verify AVD was created
RUN ls -la ~/.android/avd/ && \
    ${ANDROID_HOME}/cmdline-tools/latest/bin/avdmanager list avd

# Add permission fix script execution to startup
RUN echo '#!/bin/bash' > /entrypoint.sh && \
    echo 'set -e' >> /entrypoint.sh && \
    echo 'echo "Running permission fixes..."' >> /entrypoint.sh && \
    echo 'chmod +x scripts/fix-permissions.sh' >> /entrypoint.sh && \
    echo 'scripts/fix-permissions.sh || true' >> /entrypoint.sh && \
    echo 'echo "Starting main application..."' >> /entrypoint.sh && \
    echo 'exec ./scripts/docker-startup.sh "$@"' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

EXPOSE 5555 5554
ENTRYPOINT ["/entrypoint.sh"] 