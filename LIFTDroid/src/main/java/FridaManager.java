// import com.google.gson.Gson; // For converting list to JSON string - Temporarily commented out

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;
import java.util.regex.Pattern; // Keep for stateToApiRegexPatterns if used elsewhere
import java.util.Map; // Keep for stateToApiRegexPatterns if used elsewhere
import java.util.HashMap; // Keep for stateToApiRegexPatterns if used elsewhere

// Stub class - TODO: Implement with actual Frida interaction logic
public class FridaManager {

    private Process fridaProcess;
    private final String fridaLogFilePath; // Set via constructor or config
    private final String fridaAgentJsPath; // Path to login_agent.js
    private List<String> apiRegexOnlyPatternsForAgent; // Parsed regex strings for the agent
    private String apiRegexPatternsJsonString; // JSON string of the above list

    // This map is for ApiLogAnalyzer to map API calls to States.
    // The Frida agent itself only needs the regex strings to hook methods.
    private Map<String, Pattern> stateToApiRegexPatterns;

    public FridaManager(String stateAndApiRegexFilePath, String agentJsPath, String logPath, String unused) {
        this.fridaAgentJsPath = agentJsPath;
        this.fridaLogFilePath = logPath;
        this.apiRegexOnlyPatternsForAgent = new ArrayList<>();
        this.stateToApiRegexPatterns = new HashMap<>();
        loadRegexesFromFile(stateAndApiRegexFilePath);
        prepareJsonRegexForAgent();
    }

    private void loadRegexesFromFile(String filePath) {
        System.out.println("FridaManager: Loading regex patterns from: " + filePath);
        try (BufferedReader reader = new BufferedReader(new FileReader(filePath))) {
            String line;
            int count = 0;
            while ((line = reader.readLine()) != null) {
                line = line.trim();
                if (line.startsWith("#") || line.isEmpty()) {
                    continue;
                }
                String[] parts = line.split(":::", 2);
                if (parts.length == 2) {
                    String stateName = parts[0].trim();
                    String regexStr = parts[1].trim();
                    try {
                        stateToApiRegexPatterns.put(stateName, Pattern.compile(regexStr));
                        apiRegexOnlyPatternsForAgent.add(regexStr); // Add only the regex for the agent
                        count++;
                    } catch (Exception e) {
                        System.err.println("FridaManager: Error compiling regex for state " + stateName + ": "
                                + regexStr + " - " + e.getMessage());
                    }
                } else {
                    System.err.println("FridaManager: Skipping malformed regex line: " + line);
                }
            }
            System.out.println("FridaManager: Loaded " + count + " regex patterns for API hooking and state mapping.");
        } catch (IOException e) {
            System.err.println("FridaManager: Error reading regex patterns file: " + filePath);
            e.printStackTrace();
        }
    }

    private void prepareJsonRegexForAgent() {
        if (apiRegexOnlyPatternsForAgent != null && !apiRegexOnlyPatternsForAgent.isEmpty()) {
            // Gson gson = new Gson(); // Temporarily commented out
            // this.apiRegexPatternsJsonString = gson.toJson(apiRegexOnlyPatternsForAgent);

            // Manual JSON array construction (simple version, not robust for all edge
            // cases)
            StringBuilder sb = new StringBuilder();
            sb.append("[");
            for (int i = 0; i < apiRegexOnlyPatternsForAgent.size(); i++) {
                // Basic escaping for quotes if regex string itself contains quotes.
                // Proper JSON string escaping is more complex.
                String regex = apiRegexOnlyPatternsForAgent.get(i);
                sb.append("\"").append(regex.replace("\"", "\\\"")).append("\"");
                if (i < apiRegexOnlyPatternsForAgent.size() - 1) {
                    sb.append(",");
                }
            }
            sb.append("]");
            this.apiRegexPatternsJsonString = sb.toString();

            // System.out.println("FridaManager: Prepared JSON for agent: " +
            // apiRegexPatternsJsonString);
        } else {
            this.apiRegexPatternsJsonString = "[]"; // Empty JSON array
            System.out.println("FridaManager: No API regex patterns loaded, agent will receive empty list.");
        }
    }

    // Public getter for the JSON string, so LoginDetectTest can pass it if needed,
    // or this class can use it directly in startFrida.
    public String getApiRegexPatternsJsonString() {
        return this.apiRegexPatternsJsonString;
    }

    // For ApiLogAnalyzer
    public Map<String, Pattern> getStateToApiRegexPatterns() {
        return this.stateToApiRegexPatterns;
    }

    public void startFrida(String appPackageName) {
        if (this.apiRegexPatternsJsonString == null || this.apiRegexPatternsJsonString.equals("[]")) {
            System.err.println(
                    "FridaManager: Cannot start Frida, API regex patterns JSON string is not prepared or empty.");
            return;
        }
        System.out.println("FridaManager: Attempting to start Frida for app: " + appPackageName);
        System.out.println("FridaManager: Using agent: " + fridaAgentJsPath);
        System.out.println("FridaManager: Logging API calls to: " + fridaLogFilePath);

        try {
            // Step 1: Launch Frida and attach to the application
            // Ensure the frida_log_file is empty or ready for appending
            new File(fridaLogFilePath).delete();

            // Create the basic Frida command
            List<String> command = new ArrayList<>();
            command.add("frida");
            command.add("-U"); // USB device
            command.add("-f"); // Force spawn
            command.add(appPackageName);
            command.add("-l");
            command.add(fridaAgentJsPath);
            command.add("--no-pause");
            command.add("--runtime=v8");

            ProcessBuilder processBuilder = new ProcessBuilder(command);
            // Redirect output to our log file
            processBuilder.redirectOutput(new File(fridaLogFilePath));
            processBuilder.redirectErrorStream(true);

            fridaProcess = processBuilder.start();

            // Step 2: Wait for Frida to attach (this is crucial)
            System.out.println("FridaManager: Waiting for Frida to attach to: " + appPackageName);
            TimeUnit.SECONDS.sleep(3); // Adjust as necessary

            // Step 3: Send API regex patterns to the agent via RPC
            // Note: The actual RPC mechanism requires pyfrida or similar tools
            // For this implementation, we'll use a separate python script to establish RPC

            String escapedJson = this.apiRegexPatternsJsonString.replace("\"", "\\\"");
            List<String> rpcCommand = new ArrayList<>();
            rpcCommand.add("python3");
            rpcCommand.add("scripts/frida_controller.py");
            rpcCommand.add(appPackageName);
            rpcCommand.add(escapedJson);

            System.out.println("FridaManager: Executing RPC call to send API regex patterns");
            Process rpcProcess = new ProcessBuilder(rpcCommand).start();

            // Read output from RPC command
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(rpcProcess.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    System.out.println("RPC: " + line);
                }
            }

            int exitCode = rpcProcess.waitFor();
            if (exitCode == 0) {
                System.out
                        .println("FridaManager: RPC call successful. Agent should be initialized with regex patterns.");
            } else {
                System.err.println("FridaManager: RPC call failed with exit code: " + exitCode);
            }

            // Give Frida agent time to set up all hooks
            TimeUnit.SECONDS.sleep(2);
            System.out.println("FridaManager: Frida agent should now be monitoring API calls");

        } catch (IOException | InterruptedException e) {
            System.err.println("FridaManager: Error starting Frida process: " + e.getMessage());
            e.printStackTrace();
            if (fridaProcess != null) {
                fridaProcess.destroyForcibly();
                fridaProcess = null;
            }
        }
    }

    public void stopFrida() {
        System.out.println("FridaManager: Attempting to stop Frida process...");
        if (fridaProcess != null && fridaProcess.isAlive()) {
            fridaProcess.destroy(); // Sends SIGTERM
            try {
                if (!fridaProcess.waitFor(5, TimeUnit.SECONDS)) { // Wait for graceful exit
                    System.out.println("FridaManager: Frida process did not exit gracefully, forcing kill...");
                    fridaProcess.destroyForcibly(); // Sends SIGKILL
                }
                if (!fridaProcess.isAlive()) {
                    System.out.println("FridaManager: Frida process stopped.");
                } else {
                    System.out.println("FridaManager: Frida process may still be running after forcible stop attempt.");
                }
            } catch (InterruptedException e) {
                System.err.println("FridaManager: Interrupted while waiting for Frida process to stop.");
                fridaProcess.destroyForcibly();
                Thread.currentThread().interrupt();
            }
        }
        // Additional cleanup: Try to kill any leftover frida-server on device if issues
        // persist
        try {
            Runtime.getRuntime().exec("adb shell pkill -f \"frida-server\"");
        } catch (IOException e) {
            e.printStackTrace();
        }
        System.out.println("FridaManager: stopFrida() completed.");
    }

    public void restartFrida(String appPackageName) {
        System.out.println("FridaManager: Restarting Frida...");
        stopFrida();
        // Add a small delay to ensure ports/processes are fully released if needed
        try {
            TimeUnit.MILLISECONDS.sleep(500);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        startFrida(appPackageName);
    }
}