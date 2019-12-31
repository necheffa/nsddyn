from flask import Flask, jsonify, request
from werkzeug.security import generate_password_hash, check_password_hash
app = Flask(__name__)

# function stub - simulate reading a hashed password from some location on disk
def getPasswd(username):
    if username == "alex":
        return generate_password_hash("password")
    return None

@app.route("/")
@app.route("/index")
def home_page():
    return "This will be my homepage, it will have links to sub-pages."

# NOTE: requires mimetype to be "application/json"
# {
#   username: user
#   password: plaintext
#   ipaddr: ip-address
#   hostname: host
# }
@app.route("/api/dynupd", methods=["GET", "POST"])
def dynupd():
    # TODO: sanitize the input!
    data = request.get_json()
    username = data["username"]
    # TODO: I probably don't want to store this as a string type...but what to do?
    password = data["password"]
    ipaddr = data["ipaddr"]
    hostname = data["hostname"]

    print(username, file=sys.stderr)
    print(password, file=sys.stderr)
    print(ipaddr, file=sys.stderr)
    print(hostname, file=sys.stderr)

    hashedpw = getPasswd(username)
    if check_password_hash(hashedpw, password):
        # echo POST'd JSON data
        return jsonify(data)
    return "Error! Unknown user or password.\n"
#    return "This will be the URL for my nsddyn app."

if __name__ == "__main__":
    app.run(host="0.0.0.0")

