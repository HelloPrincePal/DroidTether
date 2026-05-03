class DroidTether < Formula
    desc "Android USB tethering for Apple Silicon Macs — no kext, no SIP changes"
    homepage "https://github.com/princePal/droidtether"
    version "0.8.7"
  
    # ARM binary (Apple Silicon — M1/M2/M3/M4)
    on_arm do
      url "https://github.com/princePal/droidtether/releases/download/v#{version}/droidtether-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_AFTER_RELEASE"
    end
  
    # Intel binary (fallback)
    on_intel do
      url "https://github.com/princePal/droidtether/releases/download/v#{version}/droidtether-darwin-amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_AFTER_RELEASE"
    end
  
    license "MIT"
  
    # libusb is required for USB device access
    depends_on "libusb"
  
    # Requires macOS 13+ (Ventura) on Apple Silicon
    depends_on macos: :ventura
  
    def install
      bin.install "droidtether"
  
      # Install default config
      (etc/"droidtether").mkpath
      (etc/"droidtether/droidtether.toml").write(default_config) unless (etc/"droidtether/droidtether.toml").exist?
  
      # Install launchd plist (will be loaded in post_install)
      (prefix/"launchd").mkpath
      (prefix/"launchd/com.princePal.droidtether.plist").write(plist_content)
    end
  
    def post_install
      # Copy plist to LaunchDaemons and load it
      system "sudo", "cp", "#{prefix}/launchd/com.princePal.droidtether.plist",
             "/Library/LaunchDaemons/com.princePal.droidtether.plist"
      system "sudo", "launchctl", "load", "/Library/LaunchDaemons/com.princePal.droidtether.plist"
    end
  
    def caveats
      <<~EOS
        DroidTether has been installed and the daemon started.
  
        Usage:
          1. Connect your Android phone via USB
          2. On your phone: Settings → Network → Hotspot & Tethering → USB Tethering → ON
          3. Internet will route through your phone automatically
  
        Logs:
          tail -f /var/log/droidtether.log
  
        Config:
          #{etc}/droidtether/droidtether.toml
  
        To stop/start the daemon:
          sudo launchctl unload /Library/LaunchDaemons/com.princePal.droidtether.plist
          sudo launchctl load /Library/LaunchDaemons/com.princePal.droidtether.plist
  
        GitHub: https://github.com/princePal/droidtether
      EOS
    end
  
    # Called by `brew uninstall`
    def uninstall_preflight
      system "sudo", "launchctl", "unload", "/Library/LaunchDaemons/com.princePal.droidtether.plist"
      system "sudo", "rm", "-f", "/Library/LaunchDaemons/com.princePal.droidtether.plist"
    end
  
    test do
      # Verify binary runs and prints version
      assert_match "droidtether v", shell_output("#{bin}/droidtether --version")
    end
  
    private
  
    def plist_content
      <<~XML
        <?xml version="1.0" encoding="UTF-8"?>
        <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
          "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
        <plist version="1.0">
        <dict>
          <key>Label</key>
          <string>com.princePal.droidtether</string>
          <key>ProgramArguments</key>
          <array>
            <string>#{opt_bin}/droidtether</string>
            <string>--config</string>
            <string>#{etc}/droidtether/droidtether.toml</string>
          </array>
          <key>RunAtLoad</key>
          <true/>
          <key>KeepAlive</key>
          <true/>
          <key>StandardOutPath</key>
          <string>/var/log/droidtether.log</string>
          <key>StandardErrorPath</key>
          <string>/var/log/droidtether.log</string>
          <key>ThrottleInterval</key>
          <integer>5</integer>
        </dict>
        </plist>
      XML
    end
  
    def default_config
      # Embedded minimal default config (full reference: repo config/default.toml)
      <<~TOML
        # DroidTether config — edit as needed
        # Full reference: https://github.com/princePal/droidtether/blob/main/config/default.toml
  
        [usb]
        poll_interval_ms = 500
  
        [rndis]
        max_transfer_size = 16384
  
        [tun]
        mtu = 1500
  
        [dhcp]
        timeout_ms = 5000
  
        [route]
        set_default_route = true
  
        [logging]
        level = "info"
        format = "text"
        file = "/var/log/droidtether.log"
      TOML
    end
  end