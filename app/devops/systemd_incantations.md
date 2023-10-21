# See what current run level is

	systemctl list-units --type=target
	systemctl status


# After putting the .service file in /etc/systemd/system/dog-tracking.service

	systemctl daemon-reload
	systemctl enable dog-tracking.service
	systemctl start dog-tracking.service


# Checking status

	systemctl status dog-tracking.service


# Looking at the service log, if not redirecting STDOUT/STDERR in .service file

	journalctl -u dog-tracking.service


# If redirecting STDOUT/STDER:

	logs are in /root/dog-tags*
