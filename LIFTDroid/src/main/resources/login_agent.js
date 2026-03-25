Java.perform(function () {
    console.log("[*] Starting login agent...");

    // Function to read regex patterns from file
    function readRegexPatterns() {
        try {
            const patterns = {};
            const patternFile = new File("/data/local/tmp/regex_patterns.txt");
            if (!patternFile.exists()) {
                console.log("[!] regex_patterns.txt not found, using default patterns");
                return getDefaultPatterns();
            }

            const reader = new BufferedReader(new FileReader(patternFile));
            let line;
            while ((line = reader.readLine()) != null) {
                line = line.trim();
                if (line && !line.startsWith("#")) {
                    const parts = line.split(":::");
                    if (parts.length === 2) {
                        const stateName = parts[0].trim();
                        const pattern = parts[1].trim();
                        patterns[stateName] = new RegExp(pattern);
                    }
                }
            }
            reader.close();
            console.log(`[*] Loaded ${Object.keys(patterns).length} patterns from regex_patterns.txt`);
            return patterns;
        } catch (e) {
            console.log(`[!] Error reading regex patterns: ${e}`);
            return getDefaultPatterns();
        }
    }

    // Default patterns as fallback
    function getDefaultPatterns() {
        return {
            InitialState: /^(onCreate|onStart|onResume)$/,
            CheckingToken: /^(validateToken|checkAuthToken)$/,
            CredentialInput: /^(showLoginScreen|displayCredentialsView)$/,
            Authenticating: /^(performLoginAction|initiateAuthentication)$/,
            CheckingCredentials: /^(AuthService\.checkCredentials|UserAuthenticator\.verifyPassword)$/,
            ServiceUnavailable_Auth: /^(onNetworkError|handleServerUnreachable)$/,
            InvalidCredentials: /^(onLoginFailure|handleIncorrectCredentials)$/,
            LoggedIn_NoMFA: /^(onLoginSuccessMfaNotRequired|proceedToAppWithoutMfa)$/,
            AppMainScreen: /^(navigateToDashboard|showHomeScreen)$/,
            UserLogout: /^(performLogout|signOutUser)$/
        };
    }

    // Load patterns from file
    const patterns = readRegexPatterns();

    // Monitor all loaded classes
    Java.enumerateLoadedClasses({
        onMatch: function (className) {
            // Check if class matches any of our patterns
            Object.entries(patterns).forEach(([patternName, pattern]) => {
                if (pattern.test(className)) {
                    try {
                        const targetClass = Java.use(className);
                        const methods = targetClass.class.getDeclaredMethods();

                        methods.forEach(function (method) {
                            const methodName = method.getName();
                            // Log the API call with timestamp and state
                            const timestamp = new Date().toISOString();
                            console.log(`[${timestamp}] [${patternName}] ${className}.${methodName}`);

                            // Hook the method
                            const overloads = targetClass[methodName].overloads;
                            overloads.forEach(function (overload) {
                                overload.implementation = function () {
                                    try {
                                        const result = this[methodName].apply(this, arguments);
                                        // Log method arguments and return value
                                        console.log(`[${timestamp}] [${patternName}] ${className}.${methodName} called with args: ${JSON.stringify(Array.from(arguments))}`);
                                        console.log(`[${timestamp}] [${patternName}] ${className}.${methodName} returned: ${result}`);
                                        return result;
                                    } catch (e) {
                                        console.log(`[ERROR] [${patternName}] ${className}.${methodName}: ${e}`);
                                        throw e;
                                    }
                                };
                            });
                        });
                    } catch (e) {
                        // Ignore errors for classes we can't hook
                    }
                }
            });
        },
        onComplete: function () {
            console.log("[*] Class enumeration completed");
        }
    });

    // Monitor common login activity patterns
    const activityPatterns = [
        'LoginActivity',
        'SignInActivity',
        'AuthActivity',
        'AuthenticationActivity'
    ];

    activityPatterns.forEach(function (activityPattern) {
        try {
            Java.enumerateLoadedClasses({
                onMatch: function (className) {
                    if (className.includes(activityPattern)) {
                        const activity = Java.use(className);
                        if (activity.onCreate) {
                            activity.onCreate.overload('android.os.Bundle').implementation = function (bundle) {
                                console.log(`[+] ${className}.onCreate() called`);
                                this.onCreate(bundle);
                            };
                        }
                    }
                },
                onComplete: function () { }
            });
        } catch (e) {
            // Ignore errors for activities we can't hook
        }
    });

    console.log("[*] Login agent started successfully");
}); 