import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

public class ApiLogAnalyzer {

    private final Map<String, Pattern> stateRegexPatterns; // Map<StateName, RegexPatternForApi>
    private final List<ApiCallEntry> apiCallLog;

    public static class ApiCallEntry {
        long timestamp;
        String className;
        String methodName;
        String mappedState; // State this API call is mapped to

        public ApiCallEntry(long timestamp, String className, String methodName) {
            this.timestamp = timestamp;
            this.className = className;
            this.methodName = methodName;
            this.mappedState = "Unknown"; // Default state
        }

        public String getMethodSignature() {
            return className + "." + methodName;
        }

        @Override
        public String toString() {
            return "ApiCallEntry{" +
                    "timestamp=" + timestamp +
                    ", className=\'" + className + "\'" +
                    ", methodName=\'" + methodName + "\'" +
                    ", mappedState=\'" + mappedState + "\'" +
                    '}';
        }
    }

    public ApiLogAnalyzer(String regexPatternsFilePath) throws IOException {
        this.stateRegexPatterns = loadRegexPatterns(regexPatternsFilePath);
        this.apiCallLog = new ArrayList<>();
    }

    private Map<String, Pattern> loadRegexPatterns(String filePath) throws IOException {
        Map<String, Pattern> patterns = new HashMap<>();
        try (BufferedReader reader = new BufferedReader(new FileReader(filePath))) {
            String line;
            while ((line = reader.readLine()) != null) {
                line = line.trim();
                if (line.startsWith("#") || line.isEmpty()) {
                    continue; // Skip comments and empty lines
                }
                String[] parts = line.split(":::", 2);
                if (parts.length == 2) {
                    try {
                        patterns.put(parts[0].trim(), Pattern.compile(parts[1].trim()));
                    } catch (Exception e) {
                        System.err.println("Error compiling regex for state " + parts[0] + ": " + parts[1] + " - "
                                + e.getMessage());
                    }
                } else {
                    System.err.println("Skipping malformed regex line: " + line);
                }
            }
        }
        return patterns;
    }

    public void parseLogFile(String logFilePath) throws IOException {
        this.apiCallLog.clear(); // Clear previous logs if any
        try (BufferedReader reader = new BufferedReader(new FileReader(logFilePath))) {
            String line;
            // Frida log format: TIMESTAMP ::: CLASS_NAME.METHOD_NAME
            Pattern logEntryPattern = Pattern.compile("([0-9]+)[ ]*:::[ ]*([^.]+).([^ ]+)");

            while ((line = reader.readLine()) != null) {
                Matcher matcher = logEntryPattern.matcher(line);
                if (matcher.find()) {
                    try {
                        long timestamp = Long.parseLong(matcher.group(1));
                        String className = matcher.group(2);
                        String methodName = matcher.group(3);
                        ApiCallEntry entry = new ApiCallEntry(timestamp, className, methodName);
                        mapApiCallToState(entry);
                        this.apiCallLog.add(entry);
                    } catch (NumberFormatException e) {
                        System.err.println("Skipping log line due to timestamp parse error: " + line);
                    } catch (Exception e) {
                        System.err
                                .println("Skipping log line due to unexpected error: " + line + " - " + e.getMessage());
                    }
                } else {
                    // Optionally log lines that don\'t match the expected Frida output,
                    // but be mindful of Frida's own console.log messages from the agent.
                    if (line.contains(" ::: ")) { // Heuristic to only log potential API calls
                        System.err.println("Skipping unparsable API log line: " + line);
                    }
                }
            }
        }
    }

    private void mapApiCallToState(ApiCallEntry entry) {
        String fullMethodName = entry.getMethodSignature();
        for (Map.Entry<String, Pattern> regexEntry : this.stateRegexPatterns.entrySet()) {
            if (regexEntry.getValue().matcher(fullMethodName).matches()
                    || regexEntry.getValue().matcher(entry.methodName).matches()) {
                // First check full signature, then just method name for broader compatibility
                entry.mappedState = regexEntry.getKey();
                return; // First match wins
            }
        }
        // If no specific regex matches, it remains "Unknown"
    }

    public List<ApiCallEntry> getApiCallLog() {
        return apiCallLog;
    }

    public List<String> getObservedStateSequence() {
        return apiCallLog.stream()
                .map(entry -> entry.mappedState)
                .filter(state -> !state.equals("Unknown")) // Optionally filter out Unknown states
                .collect(Collectors.toList());
    }

    public List<String> getObservedStateSequenceWithUnknown() {
        return apiCallLog.stream()
                .map(entry -> entry.mappedState)
                .collect(Collectors.toList());
    }

    // Placeholder for actual state machine transition validation
    public List<String> findIncorrectTransitions(List<String> observedStates,
            Map<String, List<String>> validTransitions) {
        List<String> errors = new ArrayList<>();
        if (observedStates.size() < 2) {
            return errors; // Not enough states for a transition
        }

        for (int i = 0; i < observedStates.size() - 1; i++) {
            String currentState = observedStates.get(i);
            String nextState = observedStates.get(i + 1);

            if (currentState.equals("Unknown") || nextState.equals("Unknown")) {
                // Optionally decide how to handle transitions involving Unknown states
                // For now, we might skip them or log them as potential points of interest
                // errors.add("Transition involving Unknown state: " + currentState + " -> " +
                // nextState);
                continue;
            }

            List<String> allowedNextStates = validTransitions.get(currentState);
            if (allowedNextStates == null || !allowedNextStates.contains(nextState)) {
                errors.add("Invalid transition: " + currentState + " -> " + nextState + " at index " + i);
            }
        }
        return errors;
    }

    public static void main(String[] args) {
        // Example Usage:
        try {
            // 1. Create regex_patterns.txt (as described before)
            // Example:
            // checktoken:::^(?:(?:check[_]?(?:Credentials|creds|Password|pwd))|(?:(?:validate|verify)[_]?(?:Credentials|creds|Password|pwd))|(?:(?:credentials|password)(?:Validation|_verify))|(?:auth(?:Validation|_verify)))$
            // LoginSuccess:::onLoginSuccess
            // CredentialInput:::setText

            // 2. Create a dummy frida_api_log.txt
            // Example:
            // 1678886400000 ::: com.example.app.LoginActivity.onCreate
            // 1678886401000 ::: com.example.app.CredentialsManager.setText
            // 1678886402000 ::: com.example.app.AuthService.checktoken
            // 1678886403000 ::: com.example.app.NavigationHandler.onLoginSuccess
            // 1678886403500 ::: com.example.app.SomeOtherClass.unexpectedMethod
            // 1678886404000 ::: com.example.app.AuthService.checktoken // Another
            // checktoken

            String regexFile = "regex_patterns.txt"; // Path to your regex patterns
            String logFile = "frida_api_log.txt"; // Path to the log file from Frida

            ApiLogAnalyzer analyzer = new ApiLogAnalyzer(regexFile);
            analyzer.parseLogFile(logFile);

            System.out.println("--- Raw API Call Log with Mapped States ---");
            analyzer.getApiCallLog().forEach(System.out::println);

            System.out.println("\n--- Observed State Sequence (excluding Unknown) ---");
            List<String> stateSequence = analyzer.getObservedStateSequence();
            stateSequence.forEach(System.out::println);

            System.out.println("\n--- Observed State Sequence (including Unknown) ---");
            List<String> stateSequenceWithUnknown = analyzer.getObservedStateSequenceWithUnknown();
            stateSequenceWithUnknown.forEach(System.out::println);

            // Define your valid state transitions (as per your state machine
            // diagram/definition)
            // This is crucial for the "Incorrect Navigation Flow" oracle
            Map<String, List<String>> validTransitions = new HashMap<>();
            validTransitions.put("Initial", List.of("CredentialInput", "Authenticating"));
            validTransitions.put("CredentialInput", List.of("Authenticating", "UserInputFocus", "checktoken"));
            validTransitions.put("UserInputFocus", List.of("CredentialInput"));
            validTransitions.put("Authenticating",
                    List.of("LoginSuccess", "LoginFailure", "NetworkRequestSent", "checktoken"));
            validTransitions.put("checktoken", List.of("LoginSuccess", "LoginFailure", "Authenticating"));
            validTransitions.put("LoginSuccess", List.of("Initial")); // Example: loops back or goes to main app state
            validTransitions.put("LoginFailure", List.of("CredentialInput", "Initial"));
            // ... add all valid transitions for your defined states

            System.out.println("\n--- Incorrect Navigation Flow Detection ---");
            List<String> navigationErrors = analyzer.findIncorrectTransitions(stateSequence, validTransitions);
            if (navigationErrors.isEmpty()) {
                System.out.println("No incorrect navigation flows detected.");
            } else {
                navigationErrors.forEach(System.err::println);
            }

        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}