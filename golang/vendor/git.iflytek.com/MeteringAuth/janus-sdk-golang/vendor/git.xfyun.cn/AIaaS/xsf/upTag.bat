set tag=1.4.66

git tag -d %tag%
git push origin :refs/tags/%tag%

git tag  %tag%
git push origin %tag%