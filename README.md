# lobstre

> **L**inux **OBS** **STRE**amdeck controller

This is a simple Streamdeck controller I wrote for controlling 3 `portable`
simultaneous OBS instances. This controller depends on the OBS Websocket
plugin.

Right now the controller is not really pluggable; functionality is fixed.
I'm using a Streamdeck MK.2, which has 3x5 buttons. This controller
assigns 1 row per OBS instance. For each row, button 1 toggles the live
stream on/off. Buttons 2, 3 and 4 select the first, second or third scene
defined in the OBS instance respectively. The fifth button allows me
to toggle between music (desktop audio capture) and a microphone source.

The controller listens to OBS events, so any changes/updates made in the
OBS GUI will be reflected/updated on the streamdeck device as well.

See the `config.yaml` example for configuration.

Feel free to open any issues with bugs, suggestions or pull-requests.
