from flask import Flask, redirect, render_template, request
import ldap

app = Flask(__name__)
app.logger.setLevel(20)  # log info
app.config.from_object('ghostream.default_settings')


@app.route('/')
def index():
    """Welcome page"""
    return render_template('index.html')


@app.route('/<path:path>')
def viewer(path):
    """Show stream that match this path"""
    return render_template('viewer.html', path=path)


@app.route('/app/auth', methods=['POST'])
def auth():
    """Authentication on stream start"""
    name = request.form.get('name')
    password = request.form.get('pass')

    # Stream need a name and password
    if name is None or password is None:
        # When login success, the RTMP is redirected to remove the "?pass=xxx"
        # so just ignore login here, and NGINX will still allow streaming.
        return "Malformed request", 400

    ldap_user_dn = app.config.get('LDAP_USER_DN')
    bind_dn = f"cn={name},{ldap_user_dn}"
    try:
        # Try to bind LDAP as the user
        ldap_uri = app.config.get('LDAP_URI')
        connect = ldap.initialize(ldap_uri)
        connect.bind_s(bind_dn, password)
        connect.unbind_s()
        app.logger.info("%s logged in successfully", name)
        # Remove "?pass=xxx" from RTMP URL
        return redirect(f"rtmp://127.0.0.1:1925/app/{name}", code=302)
    except Exception:
        app.logger.warning("%s failed to log in", name)
        return 'Incorrect credentials', 401
