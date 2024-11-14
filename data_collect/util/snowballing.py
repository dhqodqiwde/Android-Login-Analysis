import requests
from collections import Counter
import nltk
from nltk.corpus import stopwords

# Ensure you have the stopwords downloaded
nltk.download('stopwords')
stop_words = set(stopwords.words('english'))


def search_github_issues(keyword, repo):
    """
    Search GitHub issues based on a keyword within a specific repository.
    """
    url = f"https://api.github.com/search/issues?q={
        keyword}+in:body+repo:{repo}"
    headers = {'Accept': 'application/vnd.github.v3+json'}
    response = requests.get(url, headers=headers)
    if response.status_code == 200:
        return response.json()
    else:
        print(f"Error: {response.status_code}")
        return None


def extract_new_keywords(issues):
    """
    Extract new keywords from the issue bodies.
    """
    keywords = set()
    potential_keywords = [
        'authentication', 'authorization', 'signin', 'signup', 'oauth',
        'session', 'password', 'credential', 'access', 'refresh token',
        'jwt', 'sso', '2fa', 'mfa', 'security', 'login page', 'logout',
        'user management', 'user account', 'login failure', 'captcha',
        'recovery', 'reset password', 'account lockout', 'secure login',
        'identity verification', 'api key', 'access token', 'bearer token',
        'auth token'
    ]
    for issue in issues['items']:
        body = issue['body'].lower()
        for keyword in potential_keywords:
            if keyword in body and keyword not in stop_words:
                keywords.add(keyword)
    return keywords


def is_valid_keyword(keyword, min_length=3, max_length=20):
    """
    Check if a keyword is of valid length and not a stopword.
    """
    return min_length <= len(keyword) <= max_length and keyword not in stop_words


def filter_keywords(keywords, frequency_threshold=1):
    """
    Filter out keywords that appear less frequently than the threshold.
    """
    keyword_counter = Counter(keywords)
    return {k for k, v in keyword_counter.items() if v > frequency_threshold}


initial_keywords = ['login', 'token', 'auth']
all_keywords = set(initial_keywords)
repo = "<repository>"  # Replace with your repository name

# Initialize the set of keywords to search
keywords_to_search = set(initial_keywords)
new_keyword_count = len(keywords_to_search)

# Loop until no new keywords appear
while new_keyword_count > 0:
    new_keywords_to_search = set()
    for keyword in keywords_to_search:
        issues = search_github_issues(keyword, repo)
        if issues:
            new_keywords = extract_new_keywords(issues)
            # Filter invalid keywords
            valid_keywords = {
                kw for kw in new_keywords if is_valid_keyword(kw)}
            new_keywords_to_search.update(valid_keywords - all_keywords)
            all_keywords.update(valid_keywords)
    # Filter low-frequency keywords
    all_keywords = filter_keywords(all_keywords)
    new_keyword_count = len(new_keywords_to_search)
    keywords_to_search = new_keywords_to_search

print("Expanded Keywords:", all_keywords)
