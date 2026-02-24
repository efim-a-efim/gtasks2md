import os
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from google.auth.transport.requests import Request

SCOPES = ['https://www.googleapis.com/auth/tasks']

def authenticate(credentials_path: str = None) -> Credentials:
    """
    Handles authentication to Google APIs.
    Loads credentials from the given path, environment variable, or default local file.
    Manages token generation and refreshing.
    """
    creds = None
    token_path = 'token.json'
    
    # The file token.json stores the user's access and refresh tokens, and is
    # created automatically when the authorization flow completes for the first time.
    if os.path.exists(token_path):
        creds = Credentials.from_authorized_user_file(token_path, SCOPES)
    
    # If there are no (valid) credentials available, let the user log in.
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            if not credentials_path:
                credentials_path = os.environ.get('GOOGLE_APPLICATION_CREDENTIALS', 'credentials.json')
            
            if not os.path.exists(credentials_path):
                raise FileNotFoundError(
                    f"Credentials file not found at '{credentials_path}'. "
                    "Please provide a valid OAuth 2.0 Client ID JSON file."
                )
                
            flow = InstalledAppFlow.from_client_secrets_file(credentials_path, SCOPES)
            creds = flow.run_local_server(port=0)
            
        # Save the credentials for the next run
        with open(token_path, 'w') as token:
            token.write(creds.to_json())
            
    return creds
