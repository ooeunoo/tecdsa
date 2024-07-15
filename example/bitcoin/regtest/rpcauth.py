# Save this as rpcauth.py
import os
import sys
import hmac
import hashlib
import base64

def generate_salt(length=16):
    return base64.urlsafe_b64encode(os.urandom(length)).decode('utf-8')

def generate_hmac(salt, password):
    return hmac.new(salt.encode('utf-8'), password.encode('utf-8'), hashlib.sha256).hexdigest()

def generate_rpcauth(user, password):
    salt = generate_salt()
    hmac = generate_hmac(salt, password)
    return 'rpcauth={}:{}${}'.format(user, salt, hmac), password

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print('Usage: {} <username>'.format(sys.argv[0]))
        sys.exit(1)

    user = sys.argv[1]
    password = base64.urlsafe_b64encode(os.urandom(16)).decode('utf-8')
    rpcauth, password = generate_rpcauth(user, password)
    print('String to be appended to bitcoin.conf:')
    print(rpcauth)
    print('Your password:')
    print(password)