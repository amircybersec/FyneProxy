# FyneProxy

Proxy App Written with Fyne Cross-Platform Framework in Go Land using Outline SDK capabilities

<img width="200" alt="Screenshot 2024-02-02 at 12 13 05 PM" src="https://github.com/amircybersec/FyneProxy/assets/117060873/80d58af8-064c-4f3a-a8f5-54f0e5ccaafc">

<img width="200" alt="Screenshot 2024-02-02 at 12 13 22 PM" src="https://github.com/amircybersec/FyneProxy/assets/117060873/483d4684-c38d-4720-9c58-c55a2f183518">

<img width="200" alt="Screenshot 2024-02-02 at 12 15 59 PM" src="https://github.com/amircybersec/FyneProxy/assets/117060873/821d3ac2-2c47-4f70-ad54-a125b1b6fc17">

<img width="200" alt="Screenshot 2024-02-02 at 12 19 35 PM" src="https://github.com/amircybersec/FyneProxy/assets/117060873/35150967-b63c-4d62-869d-302496928a4e">

The goal is to support desktop platforms first (Linux, MacOS, Windows) and impprove the user experience on these platforms. Mobile platforms support is pending resolution on running the fyne app as background service and setting up system tunnel/proxy on mobile devices.

TODO:

- [x] Pull config list from HTTPS link
- [ ] fix issues with preserving UI state (e.g. button state) when switching between pages/views
- [x] Show popup when + is pressed and text entry field and paste button
- [ ] Show individual test results for each config (udp/tcp/domain name/resolver permutations) as accordion on Test Result page
- [ ] Enable Connect button only if the list is none empty and a certain config is selected
- [ ] Show [Popup](https://docs.fyne.io/api/v2.3/widget/popup.html) to report general app errors
- [ ] Setup system proxy automatically on Windows and Linux
- [ ] Add full VPN support on Linux based on Outline CLI
- [ ] Offer options in setting to listen on LAN (share tunnel with others)
- [ ] Releade app using Geoffrey
- [ ] Add [system tray](https://docs.fyne.io/explore/systray)
