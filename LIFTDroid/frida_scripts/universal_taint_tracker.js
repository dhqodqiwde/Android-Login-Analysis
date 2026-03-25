/**
 * Universal Taint Tracker for Android Applications
 */

console.log("[*] ===== Universal Taint Tracker Starting =====");

var Config = null;
var CurrentApp = null;

// Receive config from RPC
rpc.exports = {
    loadConfig: function (configJson) {
        Config = JSON.parse(configJson);
        console.log("[*] Configuration loaded successfully");
        console.log("[*] Sources: " + Config.sources.length);
        console.log("[*] Sinks: " + Config.sinks.length);
        console.log("[*] Propagation rules: " + Config.propagation.length);
        return "OK";
    },

    setAppPackage: function (packageName) {
        CurrentApp = packageName;
        console.log("[*] Target app: " + packageName);
        return "OK";
    },

    getTaintReport: function () {
        return JSON.stringify({
            taints: Array.from(TaintDB.taints.values()),
            flows: TaintDB.flows,
            violations: TaintDB.flows.filter(f => !f.isSafe)
        });
    }
};

var TaintDB = {
    taints: new Map(),
    counter: 0,
    flows: [],

    mark: function (data, info) {
        var taintId = "TAINT_" + (this.counter++);
        this.taints.set(data.toString(), {
            id: taintId,
            type: info.type || "UNKNOWN",
            source: info.source || "UNKNOWN",
            timestamp: Date.now(),
            value: data.toString().substring(0, 50),
            propagationChain: []
        });

        console.log("[TAINT_MARK] " + taintId + " (" + info.type + ") from " + info.source);
        return taintId;
    },

    isTainted: function (data) {
        if (!data) return false;
        return this.taints.has(data.toString());
    },

    getTaint: function (data) {
        return this.taints.get(data.toString());
    },

    propagate: function (sourceData, targetData, operation) {
        if (this.isTainted(sourceData)) {
            var sourceTaint = this.getTaint(sourceData);

            // Check propagation depth limit
            if (Config && Config.options.max_propagation_depth) {
                if (sourceTaint.propagationChain.length >= Config.options.max_propagation_depth) {
                    console.log("[!] Max propagation depth reached for " + sourceTaint.id);
                    return null;
                }
            }

            var newTaint = Object.assign({}, sourceTaint);
            newTaint.propagationChain = [...sourceTaint.propagationChain, operation];
            this.taints.set(targetData.toString(), newTaint);

            console.log("[TAINT_PROP] " + sourceTaint.id + " via " + operation);
            return sourceTaint.id;
        }
        return null;
    },

    recordFlow: function (taintId, sink, isSafe, severity, details) {
        var flow = {
            taintId: taintId,
            sink: sink,
            isSafe: isSafe,
            severity: severity || "UNKNOWN",
            details: details || {},
            timestamp: Date.now()
        };
        this.flows.push(flow);
        return flow;
    },

    generateReport: function () {
        var violations = this.flows.filter(f => !f.isSafe);

        console.log("\n" + "=".repeat(80));
        console.log("TAINT ANALYSIS REPORT");
        console.log("=".repeat(80));
        console.log("Total taints: " + this.counter);
        console.log("Active taints: " + this.taints.size);
        console.log("Total flows: " + this.flows.length);
        console.log("Violations: " + violations.length);

        if (violations.length > 0) {
            console.log("\n[!] SECURITY VIOLATIONS:");
            violations.forEach(function (v, idx) {
                console.log("\n  Violation #" + (idx + 1));
                console.log("    Taint ID: " + v.taintId);
                console.log("    Severity: " + v.severity);
                console.log("    Sink: " + v.sink);
                if (v.details.key) {
                    console.log("    Key: " + v.details.key);
                }
            });
        }
        console.log("=".repeat(80) + "\n");
    }
};

function setupHooksFromConfig() {
    if (!Config) {
        console.log("[!] No configuration loaded, waiting for config...");
        return;
    }

    Java.perform(function () {
        // 1. 设置污点源
        setupSources();

        // 2. 设置传播规则
        setupPropagation();

        // 3. 设置消毒器（Sanitizer）
        setupSanitizers();

        // 4. 设置Sink检测
        setupSinks();

        // 5. 设置应用特定的Hook（如果有）
        if (CurrentApp && Config.app_specific && Config.app_specific[CurrentApp]) {
            setupAppSpecificHooks();
        }

        console.log("[✓] All hooks installed successfully");
    });
}

function setupSources() {
    console.log("[*] Setting up taint sources...");

    Config.sources.forEach(function (source) {
        source.patterns.forEach(function (pattern) {
            try {
                var TargetClass = Java.use(pattern.class);
                var methodName = pattern.method;

                // Dynamic Hook
                if (TargetClass[methodName]) {
                    var originalMethod = TargetClass[methodName];

                    TargetClass[methodName].implementation = function () {
                        var result = originalMethod.apply(this, arguments);

                        //  Check conditions
                        if (checkSourceCondition(this, arguments, result, pattern.condition)) {
                            if (result != null) {
                                var resultStr = result.toString();

                                // Check minimum length
                                if (!Config.options.min_taint_value_length ||
                                    resultStr.length >= Config.options.min_taint_value_length) {

                                    TaintDB.mark(resultStr, {
                                        type: source.type,
                                        source: pattern.class + "." + methodName + "()"
                                    });
                                }
                            }
                        }

                        return result;
                    };

                    console.log("[✓] Hooked source: " + pattern.class + "." + methodName);
                }
            } catch (e) {
                console.log("[-] Could not hook source " + pattern.class + "." + methodName + ": " + e.message);
            }
        });
    });
}

function setupPropagation() {
    if (!Config.options.track_string_operations) {
        return;
    }

    console.log("[*] Setting up propagation rules...");

    Config.propagation.forEach(function (propGroup) {
        propGroup.rules.forEach(function (rule) {
            try {
                var TargetClass = Java.use(rule.class);
                var methodName = rule.method;

                if (TargetClass[methodName]) {
                    var originalMethod = TargetClass[methodName];

                    TargetClass[methodName].implementation = function () {
                        var result = originalMethod.apply(this, arguments);

                        if (rule.type === "PROPAGATE_TO_RESULT") {
                            // Propagate from this to result
                            TaintDB.propagate(this.toString(), result ? result.toString() : "",
                                rule.class + "." + methodName + "()");

                            // Also check parameters
                            for (var i = 0; i < arguments.length; i++) {
                                if (arguments[i] && typeof arguments[i].toString === 'function') {
                                    TaintDB.propagate(arguments[i].toString(), result ? result.toString() : "",
                                        rule.class + "." + methodName + "()");
                                }
                            }
                        }

                        return result;
                    };

                    console.log("[✓] Hooked propagation: " + rule.class + "." + methodName);
                }
            } catch (e) {
                console.log("[-] Could not hook propagation " + rule.class + "." + methodName + ": " + e.message);
            }
        });
    });
}

function setupSanitizers() {
    console.log("[*] Setting up sanitizers...");

    Config.sanitizers.forEach(function (sanitizerGroup) {
        sanitizerGroup.rules.forEach(function (rule) {
            try {
                var TargetClass = Java.use(rule.class);
                var methodName = rule.method;

                if (TargetClass[methodName]) {
                    var originalMethod = TargetClass[methodName];

                    TargetClass[methodName].implementation = function () {
                        var result = originalMethod.apply(this, arguments);

                        // Check if input is tainted
                        for (var i = 0; i < arguments.length; i++) {
                            if (arguments[i] && typeof arguments[i].toString === 'function') {
                                if (TaintDB.isTainted(arguments[i].toString())) {
                                    var taint = TaintDB.getTaint(arguments[i].toString());
                                    console.log("[+] SANITIZED: " + taint.id + " via " +
                                        rule.class + "." + methodName + "()");
                                }
                            }
                        }

                        return result;
                    };

                    console.log("[✓] Hooked sanitizer: " + rule.class + "." + methodName);
                }
            } catch (e) {
                console.log("[-] Could not hook sanitizer " + rule.class + "." + methodName + ": " + e.message);
            }
        });
    });
}

function setupSinks() {
    console.log("[*] Setting up sinks...");

    Config.sinks.forEach(function (sink) {
        sink.patterns.forEach(function (pattern) {
            try {
                var TargetClass = Java.use(pattern.class);
                var methodName = pattern.method;

                if (TargetClass[methodName]) {
                    var originalMethod = TargetClass[methodName];

                    TargetClass[methodName].implementation = function () {
                        // Check if specified parameter is tainted
                        var paramIndex = pattern.parameter_index || 0;

                        if (arguments[paramIndex]) {
                            var paramValue = arguments[paramIndex].toString();

                            if (TaintDB.isTainted(paramValue)) {
                                var taint = TaintDB.getTaint(paramValue);

                                // Check if encrypted
                                var isSafe = false;
                                if (pattern.check_encryption) {
                                    isSafe = isLikelyEncrypted(paramValue);
                                }

                                // Report discovery
                                console.log("\n" + "!".repeat(80));
                                console.log("[!] TAINT FLOW TO SINK DETECTED!");
                                console.log("!".repeat(80));
                                console.log("[+] Taint ID: " + taint.id);
                                console.log("[+] Taint Type: " + taint.type);
                                console.log("[+] Source: " + taint.source);
                                console.log("[+] Sink: " + pattern.class + "." + methodName + "()");
                                console.log("[+] Severity: " + sink.severity);
                                console.log("[+] Description: " + sink.description);

                                if (pattern.check_encryption) {
                                    if (!isSafe) {
                                        console.log("[!] CRITICAL: Data appears to be PLAINTEXT!");
                                    } else {
                                        console.log("[+] Data appears encrypted");
                                    }
                                }

                                if (taint.propagationChain.length > 0) {
                                    console.log("[+] Propagation: " + taint.propagationChain.join(" -> "));
                                }

                                //  Call stack
                                if (Config.options.generate_call_stack) {
                                    try {
                                        console.log("\n[+] Call stack:");
                                        console.log(Java.use("android.util.Log").getStackTraceString(
                                            Java.use("java.lang.Exception").$new()
                                        ));
                                    } catch (e) { }
                                }

                                console.log("!".repeat(80) + "\n");

                                // Record flow
                                var details = {};
                                if (paramIndex > 0 && arguments[paramIndex - 1]) {
                                    details.key = arguments[paramIndex - 1].toString();
                                }

                                TaintDB.recordFlow(taint.id,
                                    pattern.class + "." + methodName + "()",
                                    isSafe,
                                    sink.severity,
                                    details);
                            }
                        }

                        return originalMethod.apply(this, arguments);
                    };

                    console.log("[✓] Hooked sink: " + pattern.class + "." + methodName);
                }
            } catch (e) {
                console.log("[-] Could not hook sink " + pattern.class + "." + methodName + ": " + e.message);
            }
        });
    });
}

function setupAppSpecificHooks() {
    console.log("[*] Setting up app-specific hooks for " + CurrentApp);

    var appConfig = Config.app_specific[CurrentApp];

    // App-specific sources
    if (appConfig.sources) {
        appConfig.sources.forEach(function (source) {
            try {
                var TargetClass = Java.use(source.class);
                var methodName = source.method;

                if (TargetClass[methodName]) {
                    var originalMethod = TargetClass[methodName];

                    TargetClass[methodName].implementation = function () {
                        var result = originalMethod.apply(this, arguments);

                        if (result != null) {
                            TaintDB.mark(result.toString(), {
                                type: source.type,
                                source: source.class + "." + methodName + "() [APP_SPECIFIC]"
                            });
                        }

                        return result;
                    };

                    console.log("[✓] Hooked app-specific source: " + source.class + "." + methodName);
                }
            } catch (e) {
                console.log("[-] Could not hook app-specific source: " + e.message);
            }
        });
    }

    // App-specific sinks
    if (appConfig.sinks) {
        appConfig.sinks.forEach(function (sink) {
            try {
                var TargetClass = Java.use(sink.class);
                var methodName = sink.method;

                if (TargetClass[methodName]) {
                    var originalMethod = TargetClass[methodName];

                    TargetClass[methodName].implementation = function () {
                        var paramIndex = sink.parameter_index || 0;

                        if (arguments[paramIndex] && TaintDB.isTainted(arguments[paramIndex].toString())) {
                            var taint = TaintDB.getTaint(arguments[paramIndex].toString());

                            console.log("\n[!] APP-SPECIFIC SINK TRIGGERED!");
                            console.log("[+] " + sink.class + "." + methodName + "()");
                            console.log("[+] Taint: " + taint.id + " (" + taint.type + ")");
                            console.log("[+] Severity: " + sink.severity);

                            TaintDB.recordFlow(taint.id,
                                sink.class + "." + methodName + "()",
                                false,
                                sink.severity);
                        }

                        return originalMethod.apply(this, arguments);
                    };

                    console.log("[✓] Hooked app-specific sink: " + sink.class + "." + methodName);
                }
            } catch (e) {
                console.log("[-] Could not hook app-specific sink: " + e.message);
            }
        });
    }
}

function checkSourceCondition(thisObj, args, result, condition) {
    if (!condition) return true;

    try {
        var condLower = condition.toLowerCase();

        // Check inputType
        if (condLower.includes("inputtype contains password")) {
            try {
                var inputType = thisObj.getInputType();
                if ((inputType & 0x00000080) !== 0 || (inputType & 0x00000010) !== 0) {
                    return true;
                }
            } catch (e) { }
        }

        // Check hint   
        if (condLower.includes("hint contains")) {
            try {
                var hint = thisObj.getHint();
                if (hint) {
                    var hintStr = hint.toString().toLowerCase();
                    var keywords = condLower.match(/hint contains (\w+)/);
                    if (keywords && hintStr.includes(keywords[1])) {
                        return true;
                    }
                }
            } catch (e) { }
        }

        // Check parameter
        if (condLower.includes("parameter")) {
            if (args && args.length > 0 && args[0]) {
                var param = args[0].toString().toLowerCase();

                if (condLower.includes("parameter equals")) {
                    var target = condLower.match(/parameter equals '([^']+)'/);
                    if (target && param === target[1].toLowerCase()) {
                        return true;
                    }
                }

                if (condLower.includes("parameter contains")) {
                    var keyword = condLower.match(/parameter contains (\w+)/);
                    if (keyword && param.includes(keyword[1])) {
                        return true;
                    }
                }
            }
        }

        // Check value regex match
        if (condLower.includes("value matches") && result) {
            var regexMatch = condition.match(/value matches '([^']+)'/);
            if (regexMatch) {
                var regex = new RegExp(regexMatch[1]);
                return regex.test(result.toString());
            }
        }

    } catch (e) {
        console.log("[-] Error checking condition: " + e.message);
    }

    return false;
}

function isLikelyEncrypted(data) {
    if (!data || data.length < 20) return false;

    var plaintextIndicators = [
        "sessionid=", "domain=", "path=/", "password", "username",
        "{", "HttpOnly", "Secure", "expires=", "token="
    ];

    for (var i = 0; i < plaintextIndicators.length; i++) {
        if (data.indexOf(plaintextIndicators[i]) >= 0) {
            return false;
        }
    }

    var base64Regex = /^[A-Za-z0-9+\/]+=*$/;
    if (base64Regex.test(data)) {
        return false;
    }

    var entropy = calculateEntropy(data);
    return entropy >= 4.0;
}

function calculateEntropy(str) {
    var frequencies = {};
    for (var i = 0; i < str.length; i++) {
        var char = str.charAt(i);
        frequencies[char] = (frequencies[char] || 0) + 1;
    }

    var entropy = 0;
    var len = str.length;
    for (var char in frequencies) {
        var p = frequencies[char] / len;
        entropy -= p * (Math.log(p) / Math.log(2));
    }

    return entropy;
}

// ============================================
// Initialization
// ============================================

// Delay loading, waiting for config
setTimeout(function () {
    if (Config) {
        setupHooksFromConfig();

        // Regular reporting
        if (Config.options && Config.options.report_interval_seconds) {
            setInterval(function () {
                if (TaintDB.flows.length > 0) {
                    TaintDB.generateReport();
                }
            }, Config.options.report_interval_seconds * 1000);
        }
    } else {
        console.log("[!] No configuration loaded. Please call loadConfig() via RPC");
    }
}, 1000);

console.log("\n[*] Universal Taint Tracker initialized");
console.log("[*] Waiting for configuration...");
console.log("[*] Use: frida -U -f <app> -l universal_taint_tracker.js");
