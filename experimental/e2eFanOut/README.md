# Max parallelization of E2E tests with Sonobuoy

If you try and run the e2e tests in parallel you hit a bottleneck quickly because you can only turn parallel on/off
and it chooses the 'ideal' number of parallel threads to run.

The problem with this is that failures tend to stall for long periods of time (5m at a time) and so it slows down every other test which would only take a few seconds.

The solution to this is to break of the tests into numerous plugins and run each plugin.

This can be unwieldy if you have to have 40 plugins locally. As a result, we wrap the whole process in another plugin.

So when we run this plugin it will:

- Get the list of tests in the conformance-lite mode
  - This is done by running `sonobuoy e2e` with focus and skip from conformance-lite
  - Must be careful to quote/escape things properly
  - sonobuoy gen plugin e2e -m conformance-lite| yq '.spec.env[] | select(.name == "E2E_FOCUS") | .value' but in quotes
  - sonobuoy gen plugin e2e -m conformance-lite| yq '.spec.env[] | select(.name == "E2E_SKIP") | .value' but then we need that in quotes
- Then we need to split those tests into groups of 5 (# is arbitrary, trying to be consistent and only have a few)

```
rm ./tmpplugins/p*
cat tmpversions.txt|xargs -t -I % sh -c \
  'sonobuoy gen plugin e2e --plugin-env=e2e.E2E_EXTRA_ARGS= | sed "s/plugin-name: e2e/plugin-name: e2e%/" > ./tmpplugins/p%.yaml'
```

dump all those tests into a file
2 of those tests have quotes, some have *
Removing the quotes for ease and changin \n to NUL for xargs we get:

```
cat tmptestlist| tr '\n' '\0' | sed 's/\"/\*/g' | sed 's/\[/\\\[/g' | sed 's/\]/\\\]/g' | xargs -0 -n5 bash -c 'echo "$1|$2|$3|$4|$5"' bash > focusList

while read line; do 
    echo $i sonobuoy gen plugin e2e --e2e-focus=$f2 --e2e-skip= --plugin-env=e2e.E2E_PARALLEL=true
    i += 1 
done

cat focusListWithCounts | tr '\n' '\0' | xargs -0 n2 bash -c 'sonobuoy gen plugin e2e --e2e-focus=$f2 --e2e-skip= --plugin-env=e2e.E2E_PARALLEL=true > f$1' bash
```
now we just need those plugins named differently