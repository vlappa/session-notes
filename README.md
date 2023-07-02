nsm-notes: session notes for Non Session Manager  

gui library: go-fltk  
osc library: scgolang/osc (open sound control)  

Due to limitations of the scgolang/osc library, this nsmclient libary  
can't handle clients without NSM capabilities* (one could use :message: as  
workaround). And because it isn't able to handle globs in OSC addressses,  
it can't handle NSM :broadcast:.  

Work In Progress, not ready for distribution.  

* scgolang/osc doesn't seems to be able to send empty messages.
