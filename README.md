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
 too).  The URL to configure in forgejo would be <Bot URL>/forgejo.  Note that a single deployment
 sends both github and forgejo notifications to the same channel.

Failed or cancelled status checks are announced; successes are not.  On github enable the check run,
 check suite and status events.  On forgejo 17 or later enable the workflow run and workflow job
 failure and cancelled events.

Links in the notifications can be rewritten, for example to point from a private forgejo instance
 to its public github mirror.  Rules are comma separated in from=to format and applied in order to
 the outgoing messages:

```
REWRITE_LINKS: https://forgejo.example.com/=https://github.com/,https://other.example.com/=https://github.com/other
```

```
github -rewrite-links 'https://forgejo.example.com/=https://github.com/'
```

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
