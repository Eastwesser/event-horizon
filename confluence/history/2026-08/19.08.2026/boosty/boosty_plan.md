From what you copied, I do not see:

webhook settings
callback URL settings
return/success URL settings
API keys
payment secret
custom metadata / external id fields
So no, you’re not doing anything wrong. It just looks like Boosty may not expose the kind of direct payment-gateway controls we hoped for in the normal creator UI.

What I need next is much more specific. Please look for pages/screens like:

Create subscription level Need to see:
title
price
description
any custom link / redirect / automation options
Create post / paid post / bundle Need to see whether a one-time payment object can be created, and whether it has:
external link
callback
custom parameter
post-purchase redirect
Connected Apps or any developer/integration area Need to know whether there is:
API token
webhook
bot event delivery
third-party app / automation support
Payment completion behavior Very important:
after a user pays on Boosty, where do they land?
can you configure a redirect URL?
can Boosty pass any identifier back?
Right now my best read is:

Boosty is acting more like a hosted creator subscription platform
Event Horizon currently expects something closer to a payment provider callback
those two shapes may not match directly
So the likely paths are now:

Simple path: use Boosty as an external support/subscription page, and manually/admin-confirm access in EH
Middle path: use Boosty redirect + success return if Boosty supports redirects
Full path: only if Boosty actually has webhooks/API/custom metadata somewhere
If you want, the next best thing is for me to turn boosty_payment_methods.md into a short investigation checklist with “where to click next” and “what exact field names to search for.”