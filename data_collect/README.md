# Data Collection

This directory contains various scripts and input/output files related to the data collection phase of the project. Below is a breakdown of each component and its purpose.

## Structure

- **/input**: Contains seed files for initiating searches.
  - `repos_filter.txt`: List of repositories to be filtered.
  - `repos_topics.txt`: List of topics for repository search.
  - `snowballing_keywords.txt`: Keywords for extending searches via snowballing.

- **/models**: Directory for machine learning models (not detailed here).

- **/mysql**: Directory for MySQL scripts (not detailed here).

- **/output**: Contains results from data collection scripts.
  - `GuthbIssues.txt`: Contains issues fetched from GitHub.
  - `repos_keywords_filter.txt`: Repositories filtered by specific keywords.
  - `searched_repos.txt`: Repositories fetched based on topics.
  - `TF_IDF_body.txt`: TF-IDF analysis output for bodies of texts.
  - `TF_IDF.txt`: TF-IDF analysis output.
  - `words_count_body.txt`: Count of words in the bodies of texts.
  - `words_count.txt`: General word count output.

- **/util**: Utility scripts for processing data.
  - `dependencies.go`: Manages dependencies for Go scripts.
  - `filter.py`: Python script for filtering data.
  - `issues.go`: Go script to fetch issues from GitHub.
  - `output.go`: Manages output operations for Go scripts.
  - `read.go`: Reads data inputs.
  - `repositories.go`: Fetches repository data.
  - `search.go`: Script to search for repositories or issues.
  - `snowballing.py`: Extends searches based on new keywords found.
  - `tfidf.go`: Processes TF-IDF calculations.
  - `words_count.go`: Counts words in provided texts.
  - `go.mod` and `go.sum`: Go module files for managing dependencies.
  - `main.go`: Main execution script for the Go application.

## Usage Instructions

### Searching Repositories by Topic

1. Add topics to `input/repos_topics.txt`.
2. Run the `search.go` script.
3. Insert your GitHub token when prompted to authenticate API requests.
4. Results are saved in `output/searched_repos.txt`.

### Filtering Repositories

1. Define filter keywords in `input/repos_filter.txt`.
2. Run the `filter.py` script with these keywords.
3. Results are saved in `output/repos_keywords_filter.txt`.

### Searching for Issues

1. Run the `issues.go` script to fetch issues based on predefined criteria.
2. Results are stored in `output/GuthbIssues.txt`.

### TF-IDF Analysis

- Source code for the TF-IDF computations is located in `util/tfidf.go`.
