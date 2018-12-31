# AHQUProxy
Mixer remote control proxy for Allen and Heath QU series mixers written in Golang

As of the 1.95 firmware for the QU series mixers, there's a hard limit of 8 remote control devices (QU-You, QU-Pad, Mixing station pro).

As the QU-32 has more than 8 possible "mix" outputs, it's conceiveable that a band may require more than 8 remote mix controls.   

This application provides an application layer, protocol-aware proxy for the CMNet remote mixing protocol.

Additionally, it will provide a real-time display for the following:
*  Connected remote control clients
*  Remote control client mix assignments
*  Log of changes made
