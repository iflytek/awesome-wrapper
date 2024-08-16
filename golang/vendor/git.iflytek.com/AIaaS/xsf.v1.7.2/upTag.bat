set tag=v1.7.2

git tag -d %tag%
git push origin :refs/tags/%tag%

git tag  %tag%
git push origin %tag%