# Deployment

**[← Wiki Home](Home)** · [Design](Design) · [Architecture](Architecture)

How to make `day1` run automatically on first login across platforms.

The app itself handles the "show once" logic via a sentinel file. Your job is to set up a trigger that launches `day1` at login. The app exits silently if the sentinel already exists.

---

## Windows

### Option A: RunOnce Registry Key

The simplest approach. The key is automatically deleted after the program runs.

```powershell
$exe = "C:\Program Files\day1\day1.exe"
$args = "--pages-dir `"C:\Program Files\day1\pages`""
New-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\RunOnce" `
    -Name "Day1Onboarding" `
    -Value "`"$exe`" $args" `
    -PropertyType String -Force
```

### Option B: Scheduled Task (logon trigger)

More robust. The task runs at logon and can be deployed centrally via GPO.

```powershell
$action = New-ScheduledTaskAction -Execute "C:\Program Files\day1\day1.exe" `
    -Argument '--pages-dir "C:\Program Files\day1\pages"'
$trigger = New-ScheduledTaskTrigger -AtLogon
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries

Register-ScheduledTask -TaskName "Day1Onboarding" `
    -Action $action -Trigger $trigger -Settings $settings `
    -Description "First-run onboarding wizard" -RunLevel Limited
```

---

## macOS

### LaunchAgent

Create a plist in `/Library/LaunchAgents/` (system-wide) or `~/Library/LaunchAgents/` (per-user).

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.company.day1</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/day1</string>
        <string>--pages-dir</string>
        <string>/usr/local/share/day1/pages</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/day1.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/day1.err</string>
</dict>
</plist>
```

Deploy via MDM or Fleet configuration profile.

---

## Linux

### Dependencies

WebKit2GTK on Linux doesn't ship an emoji font. Install one for emoji rendering:

```bash
sudo apt-get install -y fonts-noto-color-emoji  # Debian/Ubuntu
sudo dnf install -y google-noto-color-emoji-fonts  # Fedora/RHEL
```

### XDG Autostart

Create a `.desktop` file in `/etc/xdg/autostart/` (system-wide) or `~/.config/autostart/` (per-user).

```ini
[Desktop Entry]
Type=Application
Name=Day 1 Onboarding
Exec=/usr/local/bin/day1 --pages-dir /usr/local/share/day1/pages
X-GNOME-Autostart-enabled=true
NoDisplay=true
```

Deploy via package manager (deb/rpm) post-install script:

```bash
#!/bin/bash
install -m 755 day1 /usr/local/bin/day1
install -d /usr/local/share/day1/pages
cp pages/*.md /usr/local/share/day1/pages/
cp -r pages/assets /usr/local/share/day1/pages/
install -m 644 day1.desktop /etc/xdg/autostart/day1.desktop
```
