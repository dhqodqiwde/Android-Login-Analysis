# Android-Login-Analysis
# Android-Login-Analysis

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


# 2. Data Analysis

This section involves the open coding process used to categorize the issues gathered during data collection. Samples from different stages of coding are provided to illustrate the categorization progress.

- **Author1 and Author2**: Directories contain categorized issues from two different coders. Each directory has three files named `10%.csv`, `30%.csv`, and `60%.csv`, indicating the percentage of issues analyzed to that point.
  - `10%.csv`: Contains the first 10% of issues analyzed and categorized by the respective author.
  - `30%.csv`: Contains the first 30% of issues analyzed and categorized by the respective author.
  - `60%.csv`: Contains the first 60% of issues analyzed and categorized by the respective author.

- **Final_categorization.csv**: Consolidated file containing the final categorizations agreed upon by all coders.

- **Sample Files**:
  - `sample_10.csv`, `sample_30.csv`, `sample_60.csv`: These files provide samples of issues from the 10%, 30%, and 60% coding completion stages respectively, illustrating how the open coding evolved.

# 3. Dataset Provision

The final dataset, intended for future research, is provided in both SQL and CSV formats to accommodate different types of data analysis needs.

- **CSV Files**:
  - `t_issues.csv`: Contains all the issues used in the study.
  - `t_issues_categories.csv`: Contains categorized issues.
  - `t_repos.csv`: Contains repository data used in the study.

- **SQL Files**:
  - `t_issues.sql`: SQL script for creating and populating the issues table.
  - `t_repos.sql`: SQL script for creating and populating the repository table.

This structure ensures that researchers can access the data in a format that suits their needs, whether they are performing statistical analysis or integrating the data into a database.
