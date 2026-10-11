## Github/Forgejo IRC notifier plugin

Plugin for [IRC-Bot](https://github.com/greboid/irc-bot)

Receives notifications from a [Github](https://github.com/) or [Forgejo](https://forgejo.org/) instance and outputs them to a channel.

 - go build go build github.com/greboid/irc-github/v2/cmd/github
 - docker run greboid/irc-github
  
#### Configuration

At a bare minimum you also need to give it a channel, a github secret to verify received notifications
 on and an RPC token.  You'll like also want to specify the bot host.
 
Optionally you can hide any notifications about private channels, or send them to a different channel.

Once configured the URL to configure in github would be <Bot URL>/github

For Forgejo, also set a forgejo secret (GITEA_* headers are also accepted, so this works for Gitea
 too).  The URL to configure in forgejo would be <Bot URL>/forgejo.  Note that Forgejo does not send
 check run or check suite notifications, and that a single deployment sends both github and forgejo
 notifications to the same channel.

#### Example running

```
---
version: "3.5"
service:
  bot-github:
    image: greboid/irc-github
    environment:
      RPC_HOST: bot
      RPC_TOKEN: <as configured on the bot>
      CHANNEL: #spam
      GITHUB_SECRET: cUCrb7HJ
      FORGEJO_SECRET: 9tYwm2dR
```

```
github -rpc-host bot -rpc-token <as configured on the bot> -channel #spam -github-secret cUCrb7HJ -forgejo-secret 9tYwm2dR
```
