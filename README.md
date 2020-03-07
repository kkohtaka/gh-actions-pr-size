# GitHub Action PR Size

A GitHub Action for checking Pull Request's size

## Usage

Add the following GitHub workflow to your repository.

```yaml
name: PR size check
on
  pull_request:
    types: ['opened', 'edited']
jobs:
  check_pr_size:
    runs-on: ubuntu-latest
    steps:
      - uses: docker.pkg.github.com/kkohtaka/gh-actions-pr-size/gh-actions-pr-size:latest
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

Then, the following labels will be put on your Pull Requests depending on the size of them.

| Label    | # of changed lines |
|----------|--------------------|
| size/XS  | 1 - 10             |
| size/S   | 11 - 30            |
| size/M   | 31 - 100           |
| size/L   | 101 - 500          |
| size/XL  | 501 - 1000         |
| size/XXL | 1001 -             |

## License

[MIT](./LICENSE)
