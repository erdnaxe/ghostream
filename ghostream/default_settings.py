import os

# Default configuration

# LDAP
LDAP_URI = os.environ.get('LDAP_URI') or "ldap://127.0.0.1:389"
LDAP_USER_DN = os.environ.get('LDAP_USER_DN') or "cn=users,dc=example,dc=com"

# Web page
SITE_NAME = os.environ.get('SITE_NAME') or "Ghostream"
SITE_HOSTNAME = os.environ.get('SITE_HOSTNAME') or "localhost"
FAVICON = os.environ.get('FAVICON') or "/favicon.ico"
