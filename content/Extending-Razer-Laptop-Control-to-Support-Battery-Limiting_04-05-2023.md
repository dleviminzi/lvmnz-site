## Extending Razer Laptop Control to Support Battery Limiting 
I recently purchased the Razer Blade 14. This year's model came with a much needed software improvement: battery limiting! Previous generations of this laptop suffered from battery degradation issues. The batteries would often expand. In the worst cases, the expanding battery could damage other critical components and even shatter the glass touchpad. Many of Razer's competitors had similar issues. To combat these issues most manufacturers began to allow users to stop the laptop from charging beyond a certain percentage. My basic understanding is that limiting the charge causes a shallower depth of discharge and decreases voltage stress. Going from 100% to 0% (deep discharge) is more stressful for a battery than going from 60% to 0% (relatively shallow discharge). In terms of voltage stress, a battery at 100% is at its maximum voltage which causes stress.  

In windows, you can configure battery limits using Razer's Synapse software. In my opinion, this application is terrible. It often won't even launch. When running Linux, there is a project aptly named Razer Laptop Control that aims to replicate Synapse's feature set. I found this project while trying to figure out how to stop my keyboard from constantly cycling the rainbow.  

This software is written in Rust and it is a relatively old project. Given its age, it did not include the battery limiting feature. So, my goal was simple: add the battery limiting feature!

#### Understanding the project
![Project diagram for Razer Laptop Control](/static/images/razercontrol-arch.png)

There are two parts to this project: a daemon and a command line application. The command line application sends messages to the daemon via a socket. Those messages are parsed by the daemon which will then perform actions and save state. These actions involve sending packets to an embedded usb device. The command line application is messy, but I was able to step through it and quickly developed an understanding of what I would need to add. The daemon is much more complicated and I had to learn about the hidapi before I could even get started. 

With a basic understanding established, I turned to the community. There is a discord for razer laptop control and I sought help there. After chatting with a user, I had a better understanding of how the project worked. Here is a fancy new diagram to explain what I had learned. 

![Visual description of what Razer Laptop Control does.](/static/images/synapse-flow.png)

#### Reverse engineering USB packets 
The first step in implementing the battery limiting feature would be tracing how it worked in Windows. I was pointed to this great [guide](https://github.com/openrazer/openrazer/wiki/Reverse-Engineering-USB-Protocol) on reverse engineering the packets that Synapse sends to the USB controller. In general, you want to change the setting you want to understand and then intercept whatever packets get sent. Sadly, this meant I had to use Windows for something other than playing video games.

Following the guide above, I used wireshark to look at packets as I toggled things on and off and changed various settings. I started by finding packets that I was already familiar with for the keyboard lighting, power modes, etc. Once I knew what those looked like, it was easy for me to spot the ones not yet covered by the project. By playing with the battery tab in Synapse I managed to build the following the table: 

```
| Threshold % | Starting State | Desired State | Reserved Byte | Command | Parameter |
| ----------- | -------------- | ------------- | ------------- | ------- | --------- |
| 80          | On             | Off           | 0x07          | 0x12    | 0x50      |
| 80          | Off            | On            | 0x07          | 0x12    | 0xD0      |
| 75          | On             | Off           | 0x07          | 0x12    | 0x4B      |
| 75          | Off            | On            | 0x07          | 0x12    | 0xCB      |
| 70          | On             | Off           | 0x07          | 0x12    | 0xC6      |
| 70          | Off            | On            | 0x07          | 0x12    | 0x46      |
| 65          | On             | Off           | 0x07          | 0x12    | 0xC1      |
| 65          | Off            | On            | 0x07          | 0x12    | 0x41      |
| 60          | On             | Off           | 0x07          | 0x12    | 0xBC      |
| 60          | Off            | On            | 0x07          | 0x12    | 0x3C      |
| 55          | On             | Off           | 0x07          | 0x12    | 0xB7      |
| 55          | Off            | On            | 0x07          | 0x12    | 0x37      |
| 50          | On             | Off           | 0x07          | 0x12    | 0xB2      |
| 50          | Off            | On            | 0x07          | 0x12    | 0x32      |
```

When the threshold was on and set to a value of 50, a packet with parameter `0xB2` was sent. In binary that is `10110010`. It turns out that we only need to consider the bottom seven bits when determing the actual battery threshold value. The first bit is just there to flag whether the feature is turned on or not. It's not clear to me why the off state also includes a battery threshold value. 

With the packets figured out, all that was left was writing up some code in the daemon and the cli to take a new arguments and send a new packet type. You can see that work [here](https://github.com/phush0/razer-laptop-control-no-dkms/pull/23).

Digging into this code and figuring out how it worked was an incredible experience. At the start, I really doubted whether I would be able to do this. Even after compiling my first pass at the daemon, I was surprised to see it having any effect. Yet, I managed to get the feature implemented. The community on discord was a great help and I am looking forward to adding more features and improvements in the future!

















