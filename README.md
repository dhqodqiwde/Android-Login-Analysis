# Android Login Detector

An automated security testing tool for detecting vulnerabilities in Android application login flows. The tool combines static analysis, dynamic instrumentation (Frida), UI automation, and state machine validation to identify security issues including SSL/TLS misconfigurations, session management flaws, injection vulnerabilities, and network fault handling problems.

## Features

- **Static Analysis**: Uses SootUp framework for bytecode analysis to identify login-related methods and data flows
- **Dynamic Instrumentation**: Frida-based runtime hooking of security-critical APIs
- **UI Automation**: ADB-based automated interaction with login interfaces
- **State Machine Validation**: Detects incorrect state transitions in authentication flows
- **Taint Analysis**: Tracks sensitive data flows to detect potential leaks
- **Network Chaos Testing**: Simulates network failures to test error handling
- **Multi-Layer Bug Detection**: 6 specialized detectors for common authentication vulnerabilities

## Detected Vulnerabilities

1. **SSL/TLS Issues**: Weak certificate validation, insecure protocols
2. **Cookie Security**: Missing secure/httpOnly flags, improper scope
3. **Session Management**: Unsafe in-memory session storage
4. **Injection Attacks**: SQL injection, command injection via special characters
5. **Deserialization**: Unsafe object deserialization vulnerabilities
6. **Network Fault Handling**: ANR, crashes, or state corruption under poor network conditions

## Prerequisites

- **Java**: JDK 17 or higher
- **Maven**: 3.6+
- **Python**: 3.8+
- **Android SDK**: Platform tools (adb)
- **Frida**: 16.0+
- **Android Emulator**: With snapshot support
- **Docker** (optional): For containerized execution

## Quick Start

### 1. Installation

```bash
# Clone the repository
git clone <repository-url>
cd LoginTest

# Install Python dependencies
pip3 install frida frida-tools

# Install Android SDK tools (if not already installed)
# Download from https://developer.android.com/studio/releases/platform-tools

# Compile the project
mvn clean compile
```

### 2. Configuration

#### Configure Application Under Test

Edit `src/main/resources/config.properties`:

```properties
# Application configuration
app.path=apps/your_app.apk
app.package=com.example.app
app.activity=com.example.app.MainActivity

# Emulator configuration
emulator.name=test_device
emulator.port=5554
emulator.snapshot=snap_baseline
```

#### Configure Credential Pools

Create credential files in the `pools/` directory:

**Global pools** (used for all apps):
- `pools/accounts.txt` - One username/email per line
- `pools/passwords.txt` - One password per line
- `pools/google_accounts.txt` - Google accounts
- `pools/google_passwords.txt` - Google passwords
- `pools/facebook_accounts.txt` - Facebook accounts
- `pools/facebook_passwords.txt` - Facebook passwords

**App-specific pools** (override global pools):
Create a directory `pools/<app_package_name>/` with the same file structure.

Example for `com.example.app`:
```
pools/com.example.app/
├── accounts.txt
├── passwords.txt
├── google_accounts.txt
├── google_passwords.txt
├── facebook_accounts.txt
├── facebook_passwords.txt
└── valid_pairs.csv
```

**Format for valid_pairs.csv** (optional, for known valid credentials):
```csv
username,password
user1@example.com,password123
user2@example.com,pass456
```

### 3. Setup Emulator

```bash
# Create and start emulator
emulator -avd test_device -port 5554 &

# Install the APK
adb -s emulator-5554 install apps/your_app.apk

# Manually navigate to the login page and create a snapshot
adb -s emulator-5554 emu avd snapshot save snap_baseline

# Verify snapshot
adb -s emulator-5554 emu avd snapshot list
```

### 4. Run Detection

```bash
# Full run with environment check and compilation
./scripts/run_detector.sh

# Skip environment check
./scripts/run_detector.sh --no-check

# Skip compilation (use existing classes)
./scripts/run_detector.sh --no-compile

# Environment check only
./scripts/run_detector.sh -c
```

Or run directly with Maven:

```bash
mvn compile
mvn exec:java -Dexec.mainClass="AndroidLoginDetector"
```

### 5. View Results

Results are saved in the `reports/` directory:

```
reports/
├── <app_package>/
│   ├── login_report_<timestamp>.json     # Main detection report
│   ├── api_log_<timestamp>.txt          # Frida API logs
│   ├── taint_analysis_<timestamp>.txt   # Taint analysis results
│   └── screenshots/                     # UI screenshots
└── static_analysis/
    └── <app_package>_1.json             # Static analysis results
```

## Configuration Reference

### Test Configuration

```properties
# Number of test rounds
test.rounds=100

# Timeout per round (milliseconds)
test.timeout=360000

# Test duration per round (milliseconds)
test.duration=300000

# Wait time after login attempt (seconds)
test.login.wait=10

# Snapshot recovery wait time (seconds)
test.snapshot.wait=45
```

### Frida Configuration

```properties
# Enable/disable Frida instrumentation
frida.enabled=true

# Frida agent JavaScript file
frida.script.path=src/main/resources/login_agent.js

# Frida controller Python script
frida.controller.path=scripts/frida_controller.py

# API pattern matching file
frida.patterns.path=src/main/resources/regex_patterns.txt

# Frida timeout (seconds)
frida.timeout=120
```

### Network Chaos Testing

```properties
# Enable network chaos testing
network.chaos.enabled=true

# Network state ratio (good:bad percentage)
network.good.ratio=40
network.bad.ratio=60

# State switch interval (seconds)
network.switch.interval=60

# Chaos mode: random, periodic, adaptive
network.chaos.mode=random

# Disruption methods (comma-separated)
network.disruption.methods=disable_data,slow_network

# Disruption duration range (seconds)
network.disruption.min.duration=10
network.disruption.max.duration=30
```

### UI Element Recognition

```properties
# Login button patterns
ui.login.button.patterns=login,log in,sign in,signin,submit,enter,go,continue

# Username field patterns
ui.username.field.patterns=username,user,email,login,account,user_id

# Password field patterns
ui.password.field.patterns=password,passwd,pass,pwd,secret

# Google login patterns
ui.google.login.patterns=sign in with google,continue with google,google sign in

# Facebook login patterns
ui.facebook.login.patterns=sign in with facebook,continue with facebook,facebook login

# Login activity patterns
ui.login.activity.patterns=LoginActivity,SignInActivity,AuthActivity,LoginFragment
```

### Static Analysis

```properties
# Output directory
static.analysis.output.dir=reports/static_analysis

# Output file pattern (placeholders: {apk_name}, {version})
static.analysis.output.pattern={apk_name}_{version}.json

# Version number (for testing different analysis methods)
static.analysis.version=1

# Regex patterns file
static.analysis.patterns.path=src/main/resources/regex_patterns.txt
```

## Docker Deployment

### Build and Run

```bash
# Build Docker image
docker-compose build

# Start container
docker-compose up -d

# View logs
docker-compose logs -f

# Stop container
docker-compose down
```

### Docker Configuration

The `docker-compose.yml` provides:
- Pre-configured Android SDK and emulator
- Frida and Python dependencies
- X11 forwarding for emulator display (optional)
- Volume mounts for apps, pools, and reports

## Advanced Usage

### Custom Bug Detectors

Add custom bug detectors by extending the base class:

```java
public class BugDetectorCustom extends BugDetector {
    @Override
    public void analyze(LoginReport report, List<String> apiLog) {
        // Your detection logic
        if (vulnerabilityDetected) {
            report.addBug(new Bug(
                "CUSTOM_VULN",
                "High",
                "Description",
                evidence
            ));
        }
    }
}
```

Register in `AndroidLoginDetector.java`:

```java
bugDetectors.add(new BugDetectorCustom());
```

### Modify API Hooking

Edit `src/main/resources/login_agent.js` to hook additional methods:

```javascript
// Hook custom encryption method
Java.perform(function() {
    var CustomCrypto = Java.use("com.example.CustomCrypto");
    CustomCrypto.encrypt.implementation = function(data) {
        send({type: "crypto", method: "encrypt", data: data.toString()});
        return this.encrypt(data);
    };
});
```

### Extend Taint Analysis

Modify `src/main/java/TaintAnalyzer.java` to track additional sources/sinks:

```java
// Add custom taint source
markTaint(packageName, "custom_source", valueLength);

// Parse custom violations
if (violationType.equals("CUSTOM_LEAK")) {
    // Handle custom leak
}
```

## Project Structure

```
AndroidLoginDetector/
├── src/main/java/              # Core Java code
│   ├── AndroidLoginDetector.java      # Main orchestrator
│   ├── StaticAnalyzer.java            # SootUp-based static analysis
│   ├── FridaManager.java              # Frida integration
│   ├── UIManager.java                 # ADB-based UI automation
│   ├── StateMachineManager.java       # State machine validation
│   ├── TaintAnalyzer.java             # Taint tracking
│   ├── ApiLogAnalyzer.java            # API log analysis
│   ├── NetworkChaosManager.java       # Network fault injection
│   ├── LoginReportManager.java        # Report generation
│   ├── ConfigLoader.java              # Configuration loader
│   ├── BugDetectorSSL.java            # SSL/TLS detector
│   ├── BugDetectorCookieJar.java      # Cookie security detector
│   ├── BugDetectorInMemorySession.java # Session management detector
│   ├── BugDetectorSpecialChar.java    # Injection detector
│   ├── BugDetectorUnsafeDeserialization.java # Deserialization detector
│   └── BugDetectorNetworkFault.java   # Network fault detector
├── src/main/resources/         # Configuration and scripts
│   ├── config.properties              # Main configuration
│   ├── regex_patterns.txt             # API pattern matching
│   └── login_agent.js                 # Frida agent script
├── scripts/                    # Helper scripts
│   ├── frida_controller.py            # Frida Python controller
│   ├── mark_taint_rpc.py              # Taint marking RPC
│   ├── get_taint_report_rpc.py        # Taint report RPC
│   ├── run_detector.sh                # Main run script
│   └── docker-environment-check.sh    # Environment checker
├── pools/                      # Credential pools
│   ├── accounts.txt
│   ├── passwords.txt
│   ├── google_accounts.txt
│   ├── google_passwords.txt
│   ├── facebook_accounts.txt
│   └── facebook_passwords.txt
├── apps/                       # APK files
├── logs/                       # Log files
├── reports/                    # Detection reports
├── resources/                  # Binary resources (frida-server)
├── docker-compose.yml          # Docker Compose config
├── Dockerfile                  # Docker image definition
├── pom.xml                     # Maven configuration
└── README.md                   # This file
```

## Troubleshooting

### Frida Connection Issues

```bash
# Check Frida server on device
adb shell "su -c 'ps | grep frida'"

# Restart Frida server
adb shell "su -c 'killall frida-server'"
adb push resources/frida-server /data/local/tmp/
adb shell "su -c 'chmod 755 /data/local/tmp/frida-server'"
adb shell "su -c '/data/local/tmp/frida-server &'"

# Verify Frida connection
frida-ps -U
```

### Emulator Snapshot Issues

```bash
# List available snapshots
adb -s emulator-5554 emu avd snapshot list

# Load specific snapshot
adb -s emulator-5554 emu avd snapshot load snap_baseline

# Delete corrupted snapshot
adb -s emulator-5554 emu avd snapshot delete snap_baseline

# Recreate snapshot
# 1. Manually navigate to login page
# 2. Create new snapshot
adb -s emulator-5554 emu avd snapshot save snap_baseline
```

### Compilation Errors

```bash
# Clean and rebuild
mvn clean compile

# Update dependencies
mvn dependency:resolve

# Check Java version
java -version  # Should be 17+

# Verify Maven
mvn -version
```

### Permission Denied on Reports

```bash
# Fix permissions
chmod -R u+w reports/
rm -rf reports/*
```

## Performance Tuning

- Reduce `test.rounds` for faster testing
- Decrease `test.duration` for shorter test cycles
- Disable `network.chaos.enabled` to speed up tests
- Set `screenshot.enabled=false` to reduce I/O overhead
- Use `--no-compile` flag to skip recompilation


