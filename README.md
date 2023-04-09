# inworld-terminal-chat

Basic POC made in a weekend. Basically you can chat with your Inworld character in sort of an interactive way setting your own sprites (provided they're small enough to fit in your terminal). In this example I'm using Reimu, the bestest girl from all Gensokyo.

This is not polished and probably won't be for a long time, since tcell is not a good library for projects like this, and I'll have to refactor a few things if I switch to something else.

Should work in Linux, Windows and Mac. Only tested on macOS using iTerm2 (I love you iTerm2). It may also run in Plan9 or Illumos, but if you're into those OSs then you're very cool and maybe a bit weird.

## Instructions
1. Install and start Inworld REST API [here](https://docs.inworld.ai/docs/tutorial-integrations/node-rest/)
2. Set your own DEFAULT_PLAYER_NAME and DEFAULT_SCENE_ID and DEFAULT_SERVER_ID
3. Run this program
4. Profit!!!!
5. Press ESC to quit if you get bored of your character saying "Oh I'm sorry I don't like violence or non family-friendly conversations"

## to do
- [ ] Switch to a different terminal UI library
- [ ] Ability to switch between characters
- [ ] Set session parameters as command-line arguments (This can be done in 10 mins, don't call me lazy)
- [ ] Design a UI that doesn't visually suck (Not the task for me I'm sorry)
- [ ] Use WaitGroups from the sync library to stop literally crashing the program to stop it (facepalm)
- [ ] Close the session before crashing the program (Anothe facepalm, this should be done right now actually, P1!!!!!!!)
- [x] Buy some coffee (I ran out)


![reimu](https://en.touhouwiki.net/images/thumb/c/c8/Th18Reimu.png/408px-Th18Reimu.png)