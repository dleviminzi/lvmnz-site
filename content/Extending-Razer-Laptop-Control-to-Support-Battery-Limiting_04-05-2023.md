## Extending Razer Laptop Control to Support Battery Limiting 

I recently purchased the Razer Blade 16. This year's model came with a much needed software improvement: battery limiting! Previous generations of this laptop have suffered endlessly from battery degradation issues. This would end with the batteries expanding. In the worst case, the expanding battery could damage other critical components or even shatter the glass touchpad. Many of Razer's competitors came up with a simple solution: allow users to stop the laptop from charging beyond a certain percentage. Doing this improves long term battery health because the battery spends less time full. Given that the battery life on gaming laptops is terrible anyway, there is not much downside to this solution. 

In windows, you can configure this with the unbeleivably terrible Razer Synapse software. Luckily, when running Linux, there is a project that is aptly named Razer Laptop Control. I found this project while desperately trying to figure out how to stop my keyboard from cycling the rainbow constantly.  

This software was written in Rust and at this point it is relatively old. Given its age, it does not include the battery limiting feature. So, my goal was simple: add the battery limiting feature!

#### Understanding the project

![Project diagram for Razer Laptop Control](/static/images/razercontrol-arch.png)

There are two main parts to this project: a daemon and a command line application. The command line application sends messages to the daemon via a socket. Those messages are then parsed by the daemon which will then perform actions. Most of these actions involve sending packets to an embedded usb device. The command line application is messy, but understandable. The daemon is where things get much more complicated. 

I was off to a good start by building up an understanding of the project as it was. However, I still hadn't really done anything, so I figured it was time to figure out a first step to take. There is a discord for razer laptop control that I was hanging out in. After chatting with a user there, I learned a bit more about what was happening under the hood. Here is a fancy new diagram to explain what I had learned. 

![Visual description of what Razer Laptop Control does.](/static/images/synapse-flow.png)

#### Reverse engineering USB packets 

After diving into the project more, it became apparent that the first thing to do was figure out the packet format for the new feature. To do that, I had to [reverse engineer the packets that Synapse sends to the USB controller](https://github.com/openrazer/openrazer/wiki/Reverse-Engineering-USB-Protocol). The general idea is to intercept the usb packet sent while I changed my battery threshold limit in Synapse on Windows. Sadly, this meant I had to use Windows for something other than playing video games.

Following the guide above, I used wireshark to look at packets as I toggled things on and off and changed settings. I started by finding packets that I was already familiar with for the keyboard lighting, power modes, etc. Once I knew what those looked like, it was easy for me to spot the ones not yet covered in the project. By playing with the battery tab in Synapse I managed to build the following the table: 

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

When the threshold was on and set to a value of 50, a packet with parameter `0xB2` was sent. In binary that is `10110010` or 178 in decimal form. It turns out that we only needed to consider the bottom seven bits when determing the actual battery threshold value. The first bit was just there to flag whether the feature was turned on or not. It's not clear to me why the off state also includes a battery threshold value. 

With the packets figured out, all that was left was writing up some code in the daemon and the cli to take a new arguments and send a new packet type. You can see that work [here](https://github.com/phush0/razer-laptop-control-no-dkms/pull/23).

 Digging into this code and figuring out how it all worked was an incredible experience. At the start, I really doubted whether I would be able to do this. Even after compiling my first pass at the daemon, I was surprised to see it having any effect. Yet, I managed to get the feature implemented. The community on discord was a great help and I am looking forward to adding more features and improvements in the future!

















