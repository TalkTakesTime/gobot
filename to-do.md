To-do List
==========

### High priority

- handle login cases such as heavy load, failed logins, 522 errors, etc. See
https://github.com/TalkTakesTime/Pokemon-Showdown-Bot/blob/master/parser.js#L99-L133
- handle failed connections
- handle dropped connections
  - have a look at `net#DialTimeout` and `websocket#NewClient`?
- create directory to store log files

### Medium priority

- implement data and battling -- data handling in external repo
- expand types of githooks that can be received -- requires expanding `hookserve`
- make .git more intelligent -- low-medium priority
- decompose .git -- low-medium priority

### Low priority

- add colours to logging if logging to stdout
- add more commands
