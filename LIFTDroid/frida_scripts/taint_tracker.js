/**
 * Lightweight Taint Tracker for Android Login Security
 * 
 * 追踪敏感数据（密码、Cookie）从输入到存储的完整流程
 * 检测明文存储、不安全传输等漏洞
 */

console.log("[*] ===== Taint Tracker Starting =====");

// ============================================
// 污点追踪数据结构
// ============================================

var TaintDB = {
    // 存储所有污点数据
    taints: new Map(),

    // 污点计数器
    counter: 0,

    // 污点流记录
    flows: [],

    // 添加污点
    mark: function (data, info) {
        var taintId = "TAINT_" + (this.counter++);
        this.taints.set(data.toString(), {
            id: taintId,
            type: info.type || "UNKNOWN",
            source: info.source || "UNKNOWN",
            timestamp: Date.now(),
            value: data.toString().substring(0, 50), // 只记录前50个字符
            propagationChain: []
        });

        console.log("[TAINT_MARK] " + taintId + " (" + info.type + ") from " + info.source);
        return taintId;
    },

    // 检查是否被污染
    isTainted: function (data) {
        return this.taints.has(data.toString());
    },

    // 获取污点信息
    getTaint: function (data) {
        return this.taints.get(data.toString());
    },

    // 传播污点
    propagate: function (sourceData, targetData, operation) {
        if (this.isTainted(sourceData)) {
            var sourceTaint = this.getTaint(sourceData);
            var newTaint = Object.assign({}, sourceTaint);
            newTaint.propagationChain = [...sourceTaint.propagationChain, operation];
            this.taints.set(targetData.toString(), newTaint);

            console.log("[TAINT_PROP] " + sourceTaint.id + " via " + operation);
            return sourceTaint.id;
        }
        return null;
    },

    // 记录污点流
    recordFlow: function (taintId, sink, isSafe) {
        var flow = {
            taintId: taintId,
            sink: sink,
            isSafe: isSafe,
            timestamp: Date.now()
        };
        this.flows.push(flow);
        return flow;
    },

    // 生成报告
    generateReport: function () {
        console.log("\n" + "=".repeat(80));
        console.log("TAINT ANALYSIS REPORT");
        console.log("=".repeat(80));
        console.log("Total taints marked: " + this.counter);
        console.log("Active taints: " + this.taints.size);
        console.log("Flows detected: " + this.flows.length);

        var violations = this.flows.filter(f => !f.isSafe);
        if (violations.length > 0) {
            console.log("\n[!] SECURITY VIOLATIONS: " + violations.length);
            violations.forEach(function (v, idx) {
                console.log("\nViolation #" + (idx + 1));
                console.log("  Taint ID: " + v.taintId);
                console.log("  Sink: " + v.sink);
                console.log("  Time: " + new Date(v.timestamp).toISOString());
            });
        } else {
            console.log("\n[+] No security violations detected");
        }
        console.log("=".repeat(80) + "\n");
    }
};

// ============================================
// 污点源1: EditText密码输入
// ============================================

Java.perform(function () {
    console.log("[*] Hooking EditText for password detection...");

    try {
        var EditText = Java.use("android.widget.EditText");

        // Hook getText() - 捕获密码输入
        EditText.getText.implementation = function () {
            var text = this.getText();
            var textStr = text.toString();

            // 检查是否是密码字段
            try {
                var inputType = this.getInputType();
                var hint = this.getHint();
                var hintStr = hint ? hint.toString() : "";

                // 密码字段判断
                var isPassword = (inputType & 0x00000080) !== 0 || // TYPE_TEXT_VARIATION_PASSWORD
                    (inputType & 0x00000010) !== 0 || // TYPE_NUMBER_VARIATION_PASSWORD
                    hintStr.toLowerCase().includes("password") ||
                    hintStr.toLowerCase().includes("pwd");

                if (isPassword && textStr.length > 0) {
                    TaintDB.mark(textStr, {
                        type: "PASSWORD",
                        source: "EditText.getText()"
                    });
                }
            } catch (e) {
                // Ignore errors
            }

            return text;
        };

        console.log("[✓] EditText hook installed");
    } catch (e) {
        console.log("[-] Could not hook EditText: " + e);
    }
});

// ============================================
// 污点源2: Intent传递（密码通过Intent传递）
// ============================================

Java.perform(function () {
    console.log("[*] Hooking Intent.getStringExtra()...");

    try {
        var Intent = Java.use("android.content.Intent");

        Intent.getStringExtra.implementation = function (key) {
            var value = this.getStringExtra(key);

            if (value != null) {
                // 检查key是否表示敏感数据
                var keyLower = key.toLowerCase();
                if (keyLower.includes("password") || keyLower.includes("pwd")) {
                    TaintDB.mark(value, {
                        type: "PASSWORD",
                        source: "Intent.getStringExtra('" + key + "')"
                    });
                } else if (keyLower.includes("username") || keyLower.includes("user")) {
                    // 用户名也标记（用于关联）
                    TaintDB.mark(value, {
                        type: "USERNAME",
                        source: "Intent.getStringExtra('" + key + "')"
                    });
                }
            }

            return value;
        };

        console.log("[✓] Intent hook installed");
    } catch (e) {
        console.log("[-] Could not hook Intent: " + e);
    }
});

// ============================================
// 污点源3: HTTP响应Cookie
// ============================================

Java.perform(function () {
    console.log("[*] Hooking HTTP cookie handling...");

    try {
        // Hook OkHttp Response
        var Response = Java.use("okhttp3.Response");

        Response.header.overload('java.lang.String').implementation = function (name) {
            var value = this.header(name);

            if (value != null && name.toLowerCase() === "set-cookie") {
                TaintDB.mark(value, {
                    type: "COOKIE",
                    source: "HTTP Response (Set-Cookie header)"
                });
            }

            return value;
        };

        console.log("[✓] OkHttp Response hook installed");
    } catch (e) {
        console.log("[-] Could not hook OkHttp Response: " + e);
    }

    // NewsBlur specific: APIResponse.getCookie()
    try {
        var APIResponse = Java.use("com.newsblur.network.APIResponse");

        APIResponse.getCookie.implementation = function () {
            var cookie = this.getCookie();

            if (cookie != null && cookie.length() > 0) {
                TaintDB.mark(cookie, {
                    type: "SESSION_COOKIE",
                    source: "APIResponse.getCookie()"
                });
            }

            return cookie;
        };

        console.log("[✓] APIResponse.getCookie() hook installed");
    } catch (e) {
        // Not NewsBlur app, skip
    }
});

// ============================================
// 污点传播: 字符串操作
// ============================================

Java.perform(function () {
    console.log("[*] Hooking String operations for taint propagation...");

    try {
        var JavaString = Java.use("java.lang.String");

        // substring
        JavaString.substring.overload('int').implementation = function (start) {
            var result = this.substring(start);
            TaintDB.propagate(this.toString(), result, "String.substring()");
            return result;
        };

        // concat
        JavaString.concat.implementation = function (str) {
            var result = this.concat(str);
            if (TaintDB.isTainted(this.toString()) || TaintDB.isTainted(str)) {
                TaintDB.propagate(this.toString(), result, "String.concat()");
            }
            return result;
        };

        console.log("[✓] String operations hooks installed");
    } catch (e) {
        console.log("[-] Could not hook String operations: " + e);
    }
});

// ============================================
// 危险Sink 1: SharedPreferences明文存储
// ============================================

Java.perform(function () {
    console.log("[*] Hooking SharedPreferences for plaintext storage detection...");

    try {
        var Editor = Java.use("android.content.SharedPreferences$Editor");

        Editor.putString.implementation = function (key, value) {
            // 检查value是否被污染
            if (TaintDB.isTainted(value)) {
                var taint = TaintDB.getTaint(value);

                console.log("\n" + "!".repeat(80));
                console.log("[!] TAINT FLOW TO SINK DETECTED!");
                console.log("!".repeat(80));
                console.log("[+] Taint ID: " + taint.id);
                console.log("[+] Taint Type: " + taint.type);
                console.log("[+] Source: " + taint.source);
                console.log("[+] Propagation chain: " + taint.propagationChain.join(" -> "));
                console.log("[+] Sink: SharedPreferences.putString()");
                console.log("[+] Key: " + key);
                console.log("[+] Value length: " + value.length());

                // 检查是否加密
                var encrypted = isLikelyEncrypted(value);
                var isSafe = encrypted;

                if (!encrypted) {
                    console.log("[!] CRITICAL: Tainted data appears to be PLAINTEXT!");
                    console.log("[!] Vulnerability: Sensitive data exposure via SharedPreferences");
                    console.log("[!] Risk: Data accessible via ADB backup or root access");
                } else {
                    console.log("[+] Data appears encrypted (verify manually)");
                }

                // 打印调用栈
                console.log("\n[+] Call stack:");
                try {
                    console.log(Java.use("android.util.Log").getStackTraceString(
                        Java.use("java.lang.Exception").$new()
                    ));
                } catch (e) {
                    // Ignore
                }

                console.log("!".repeat(80) + "\n");

                // 记录流
                TaintDB.recordFlow(taint.id, "SharedPreferences.putString()", isSafe);
            }

            return this.putString(key, value);
        };

        console.log("[✓] SharedPreferences hook installed");
    } catch (e) {
        console.log("[-] Could not hook SharedPreferences: " + e);
    }
});

// ============================================
// 危险Sink 2: 文件写入
// ============================================

Java.perform(function () {
    console.log("[*] Hooking file operations...");

    try {
        var FileOutputStream = Java.use("java.io.FileOutputStream");

        FileOutputStream.write.overload('[B').implementation = function (buffer) {
            try {
                var data = Java.use("java.lang.String").$new(buffer);

                if (TaintDB.isTainted(data)) {
                    var taint = TaintDB.getTaint(data);
                    console.log("[!] Tainted data written to file!");
                    console.log("[+] Taint ID: " + taint.id);
                    console.log("[+] Type: " + taint.type);

                    TaintDB.recordFlow(taint.id, "FileOutputStream.write()", false);
                }
            } catch (e) {
                // Ignore conversion errors
            }

            return this.write(buffer);
        };

        console.log("[✓] FileOutputStream hook installed");
    } catch (e) {
        console.log("[-] Could not hook FileOutputStream: " + e);
    }
});

// ============================================
// 安全Sink: 加密存储（正例）
// ============================================

Java.perform(function () {
    console.log("[*] Hooking encryption operations...");

    try {
        var Cipher = Java.use("javax.crypto.Cipher");

        Cipher.doFinal.overload('[B').implementation = function (input) {
            try {
                var inputStr = Java.use("java.lang.String").$new(input);

                if (TaintDB.isTainted(inputStr)) {
                    var taint = TaintDB.getTaint(inputStr);
                    console.log("[+] GOOD: Tainted data being encrypted");
                    console.log("[+] Taint ID: " + taint.id);
                    console.log("[+] Type: " + taint.type);

                    // 加密后的数据不再被认为是污点（已安全处理）
                    var result = this.doFinal(input);
                    console.log("[+] Encryption completed, data is now safe");

                    return result;
                }
            } catch (e) {
                // Ignore
            }

            return this.doFinal(input);
        };

        console.log("[✓] Cipher hook installed");
    } catch (e) {
        console.log("[-] Could not hook Cipher: " + e);
    }
});

// ============================================
// 辅助函数
// ============================================

function isLikelyEncrypted(data) {
    if (!data || data.length < 20) {
        return false;
    }

    // 明显的明文特征
    var plaintextIndicators = [
        "sessionid=",
        "domain=",
        "path=/",
        "password",
        "username",
        "{",
        "HttpOnly",
        "Secure",
        "expires="
    ];

    for (var i = 0; i < plaintextIndicators.length; i++) {
        if (data.indexOf(plaintextIndicators[i]) >= 0) {
            return false;
        }
    }

    // Base64编码不等于加密
    var base64Regex = /^[A-Za-z0-9+\/]+=*$/;
    if (base64Regex.test(data)) {
        console.log("[!] WARNING: Data is Base64 encoded but may not be encrypted!");
        return false;
    }

    // 计算熵（加密数据通常有高熵）
    var entropy = calculateEntropy(data);
    if (entropy < 4.0) {
        return false; // 低熵 = 可能是明文
    }

    return true; // 可能是加密的（需人工验证）
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
// 定期报告
// ============================================

setInterval(function () {
    if (TaintDB.flows.length > 0) {
        TaintDB.generateReport();
    }
}, 30000); // 每30秒

// ============================================
// 退出时生成最终报告
// ============================================

console.log("\n[*] Taint Tracker is active!");
console.log("[*] Monitoring:");
console.log("  - Password inputs (EditText)");
console.log("  - Intent data passing");
console.log("  - HTTP cookies");
console.log("  - SharedPreferences storage");
console.log("  - File operations");
console.log("[*] Perform login to trigger taint tracking\n");
