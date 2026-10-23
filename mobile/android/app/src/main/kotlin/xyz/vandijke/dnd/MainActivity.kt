package xyz.vandijke.dnd

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.webkit.CookieManager
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import java.io.File
import java.util.UUID

class MainActivity : FlutterActivity() {
    private val channelName = "xyz.vandijke.dnd/cookies"
    private val inkChannelName = "xyz.vandijke.dnd/ink"

    // Pages shared into Villum, held until Dart asks for them. Content URIs are
    // useless to Dart and can expire, so they are copied to app storage first.
    private val sharedUris = mutableListOf<Uri>()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        captureSharedIntent(intent)
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        captureSharedIntent(intent)
    }

    @Suppress("DEPRECATION")
    private fun captureSharedIntent(intent: Intent?) {
        if (intent == null) return
        when (intent.action) {
            Intent.ACTION_SEND -> {
                val uri = intent.getParcelableExtra<Uri>(Intent.EXTRA_STREAM)
                if (uri != null && !sharedUris.contains(uri)) sharedUris.add(uri)
            }
            Intent.ACTION_SEND_MULTIPLE -> {
                val uris = intent.getParcelableArrayListExtra<Uri>(Intent.EXTRA_STREAM)
                uris?.forEach { uri -> if (!sharedUris.contains(uri)) sharedUris.add(uri) }
            }
        }
    }

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)

        // The server's `session` cookie is HttpOnly, so the OIDC flow in the
        // WebView cannot read it from JavaScript. This channel exposes the
        // native WebView cookie store to Dart instead.
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, channelName)
            .setMethodCallHandler { call, result ->
                when (call.method) {
                    "getCookies" -> {
                        val url = call.argument<String>("url")
                        if (url.isNullOrEmpty()) {
                            result.error("invalid_url", "A url argument is required", null)
                        } else {
                            result.success(CookieManager.getInstance().getCookie(url))
                        }
                    }
                    "clearCookies" -> {
                        CookieManager.getInstance().removeAllCookies(null)
                        CookieManager.getInstance().flush()
                        result.success(null)
                    }
                    else -> result.notImplemented()
                }
            }

        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, inkChannelName)
            .setMethodCallHandler { call, result ->
                when (call.method) {
                    "takeSharedImages" -> result.success(copySharedImages())
                    "readPending" -> result.success(readPending())
                    "writePending" -> {
                        val json = call.argument<String>("json")
                        if (json == null) {
                            result.error("invalid_args", "json is required", null)
                        } else {
                            writePending(json)
                            result.success(null)
                        }
                    }
                    else -> result.notImplemented()
                }
            }
    }

    private fun copySharedImages(): List<String> {
        if (sharedUris.isEmpty()) return emptyList()
        val dir = File(filesDir, "ink_inbox")
        if (!dir.exists()) dir.mkdirs()
        val paths = mutableListOf<String>()
        for (uri in sharedUris.toList()) {
            try {
                val ext = contentResolver.getType(uri)?.substringAfter("/", "png") ?: "png"
                val target = File(dir, "${UUID.randomUUID()}.$ext")
                contentResolver.openInputStream(uri)?.use { input ->
                    target.outputStream().use { output -> input.copyTo(output) }
                }
                if (target.exists() && target.length() > 0) {
                    paths.add(target.absolutePath)
                } else {
                    target.delete()
                }
            } catch (_: Exception) {
                // A provider can revoke access at any time; skip that page.
            }
        }
        sharedUris.clear()
        return paths
    }

    private fun pendingFile(): File = File(filesDir, "ink_pending.json")

    private fun readPending(): String? {
        val file = pendingFile()
        return if (file.exists()) file.readText() else null
    }

    private fun writePending(json: String) {
        pendingFile().writeText(json)
    }
}
