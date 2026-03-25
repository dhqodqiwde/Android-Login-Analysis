import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;

/**
 * Manages the reporting of login-related issues detected during testing.
 * This class centralizes the logging and report generation for various login
 * issues.
 */
public class LoginReportManager {

    private final String reportDirectory;
    private final String reportFilePath;
    private final List<IssueReport> issues;
    private final SimpleDateFormat dateFormat;

    /**
     * Represents a detected login issue with relevant details.
     */
    public static class IssueReport {
        public enum IssueType {
            NAVIGATION_FLOW_ERROR,
            CRASH,
            ANR,
            LOGIN_TIMEOUT,
            CREDENTIAL_REJECTION,
            OTHER
        }

        private final IssueType type;
        private final String description;
        private final String details;
        private final long timestamp;
        private final String credentialsUsed;

        public IssueReport(IssueType type, String description, String details, String credentialsUsed) {
            this.type = type;
            this.description = description;
            this.details = details;
            this.timestamp = System.currentTimeMillis();
            this.credentialsUsed = credentialsUsed;
        }

        @Override
        public String toString() {
            SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss.SSS");
            return String.format("[%s] %s: %s\nDetails: %s\nCredentials: %s",
                    sdf.format(new Date(timestamp)),
                    type.name(),
                    description,
                    details,
                    credentialsUsed);
        }
    }

    public LoginReportManager() {
        this("reports");
    }

    public LoginReportManager(String reportDir) {
        this.reportDirectory = reportDir;
        dateFormat = new SimpleDateFormat("yyyy-MM-dd_HH-mm-ss");
        String timestamp = dateFormat.format(new Date());
        this.reportFilePath = reportDirectory + File.separator + "login_test_report_" + timestamp + ".txt";
        this.issues = new ArrayList<>();

        // Ensure the report directory exists
        File dir = new File(reportDirectory);
        if (!dir.exists()) {
            dir.mkdirs();
        }

        // Initialize the report file with a header
        try (BufferedWriter writer = new BufferedWriter(new FileWriter(reportFilePath))) {
            writer.write("=== Login Testing Report ===\n");
            writer.write("Generated: " + dateFormat.format(new Date()) + "\n\n");
            writer.write("Issues detected will be logged below:\n");
            writer.write("==============================\n\n");
        } catch (IOException e) {
            System.err.println("Error initializing report file: " + e.getMessage());
        }
    }

    /**
     * Reports a navigation flow error
     */
    public void reportNavigationFlowError(String invalidTransition, String stateSequence, String credentials) {
        IssueReport report = new IssueReport(
                IssueReport.IssueType.NAVIGATION_FLOW_ERROR,
                "Invalid state transition detected",
                "Invalid transition: " + invalidTransition + "\nFull sequence: " + stateSequence,
                credentials);
        logIssue(report);
    }

    /**
     * Reports a crash detected in the app
     */
    public void reportCrash(String crashInfo, String credentials) {
        IssueReport report = new IssueReport(
                IssueReport.IssueType.CRASH,
                "Application crash detected",
                crashInfo,
                credentials);
        logIssue(report);
    }

    /**
     * Reports an ANR (Application Not Responding) event
     */
    public void reportANR(String anrInfo, String credentials) {
        IssueReport report = new IssueReport(
                IssueReport.IssueType.ANR,
                "ANR detected",
                anrInfo,
                credentials);
        logIssue(report);
    }

    /**
     * Reports a login timeout incident
     */
    public void reportLoginTimeout(long duration, String credentials) {
        IssueReport report = new IssueReport(
                IssueReport.IssueType.LOGIN_TIMEOUT,
                "Login process timed out",
                "Duration: " + duration + "ms",
                credentials);
        logIssue(report);
    }

    /**
     * Reports a credential rejection issue (especially with special characters)
     */
    public void reportCredentialRejection(String credentialInfo, boolean containsSpecialChars) {
        IssueReport report = new IssueReport(
                IssueReport.IssueType.CREDENTIAL_REJECTION,
                "Credentials rejected",
                "Contains special characters: " + containsSpecialChars,
                credentialInfo);
        logIssue(report);
    }

    /**
     * Reports any other type of issue
     */
    public void reportOtherIssue(String description, String details, String credentials) {
        IssueReport report = new IssueReport(
                IssueReport.IssueType.OTHER,
                description,
                details,
                credentials);
        logIssue(report);
    }

    /**
     * Logs the issue to both the console and the report file
     */
    private void logIssue(IssueReport issue) {
        // Add to the in-memory list
        issues.add(issue);

        // Print to console
        System.err.println("LOGIN BUG DETECTED: " + issue.description);
        System.err.println(issue);

        // Write to the report file
        try (BufferedWriter writer = new BufferedWriter(new FileWriter(reportFilePath, true))) {
            writer.write(issue.toString());
            writer.write("\n\n---\n\n");
        } catch (IOException e) {
            System.err.println("Error writing to report file: " + e.getMessage());
        }
    }

    /**
     * Writes test summary statistics to the report file
     */
    public void writeTestSummary(int totalAttempts, int successfulLogins, int failedLogins) {
        try (BufferedWriter writer = new BufferedWriter(new FileWriter(reportFilePath, true))) {
            writer.write("\n=== Test Summary ===\n");
            writer.write("Total login attempts: " + totalAttempts + "\n");
            writer.write("Successful logins: " + successfulLogins + "\n");
            writer.write("Failed logins: " + failedLogins + "\n");
            writer.write("Total issues detected: " + issues.size() + "\n");

            // Breakdown by issue type
            writer.write("\nIssue breakdown:\n");
            int navErrors = countIssuesByType(IssueReport.IssueType.NAVIGATION_FLOW_ERROR);
            int crashes = countIssuesByType(IssueReport.IssueType.CRASH);
            int anrs = countIssuesByType(IssueReport.IssueType.ANR);
            int timeouts = countIssuesByType(IssueReport.IssueType.LOGIN_TIMEOUT);
            int rejections = countIssuesByType(IssueReport.IssueType.CREDENTIAL_REJECTION);
            int others = countIssuesByType(IssueReport.IssueType.OTHER);

            writer.write("- Navigation flow errors: " + navErrors + "\n");
            writer.write("- Crashes: " + crashes + "\n");
            writer.write("- ANRs: " + anrs + "\n");
            writer.write("- Login timeouts: " + timeouts + "\n");
            writer.write("- Credential rejections: " + rejections + "\n");
            writer.write("- Other issues: " + others + "\n");

            writer.write("\n=== End of Report ===\n");
        } catch (IOException e) {
            System.err.println("Error writing test summary to report: " + e.getMessage());
        }
    }

    private int countIssuesByType(IssueReport.IssueType type) {
        int count = 0;
        for (IssueReport issue : issues) {
            if (issue.type == type) {
                count++;
            }
        }
        return count;
    }

    /**
     * Returns all detected issues
     */
    public List<IssueReport> getIssues() {
        return new ArrayList<>(issues);
    }

    /**
     * Get the path to the report file
     */
    public String getReportFilePath() {
        return reportFilePath;
    }
}