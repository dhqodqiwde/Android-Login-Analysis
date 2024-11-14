# Paths can be modified according to actual requirements
# Path to the first file containing keywords
path_keywords = '../input/repos_filter.txt'
# Path to the second file that needs to be filtered
path_input = '../output/searched_repos.txt'
# Path to the file where results will be written
path_output = '../output/repos_keywords_filter.txt'

# Read the keywords from the first file, strip whitespace from ends, and store them in a list
with open(path_keywords, 'r') as file:
    keywords = [line.strip().lower() for line in file]

# Read the second file and filter out lines containing any of the keywords
filtered_lines = []
with open(path_input, 'r') as file:
    for line in file:
        if not any(keyword in line.lower() for keyword in keywords):
            filtered_lines.append(line)

# Write the filtered content to the specified output file
with open(path_output, 'w') as file:
    file.writelines(filtered_lines)

print("Filtering complete, output file saved to:", path_output)
