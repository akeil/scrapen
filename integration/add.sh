DIR=./cases
name=$1
url=$2

yaml="$DIR/$name.yaml"

if [ -f "$yaml" ]; then
    echo "File $yaml already exists"
    exit;
fi



echo "url: $url" > "$yaml"

echo "# List of strings that should be PRESENT in the output" >> "$yaml"
echo "find:" >> "$yaml"
echo "  - Example" >> "$yaml"

echo "# List of strings that should NOT appear in the output" >> "$yaml"
echo "findnot:" >> "$yaml"
echo "  - Example" >> "$yaml"

echo "# List of CSS selectors and how often they are expected to appear" >> $yaml
echo "query:" >> "$yaml"
echo "  - q: nav ol li" >> "$yaml"
echo "    t: optional text content" >> "$yaml"
echo "    n: 0" >> "$yaml"
echo "  - q: img" >> "$yaml"
echo "    n: 2" >> "$yaml"

wget --output-document="$DIR/$name.html" "$url"
