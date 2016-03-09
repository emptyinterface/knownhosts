#Example

Toy example of a command line tool to run a command on a remote server over ssh.

###Build
`go build`

###Usage

	./example 10.0.0.1:22 whoami
	  -h string
	    	known hosts to verify against (default "$HOME/.ssh/known_hosts")
	  -i string
	    	key to use (default "$HOME/.ssh/id_rsa")
	  -u string
	    	user to connect as (default $USER)

LICENSE: MIT 2016