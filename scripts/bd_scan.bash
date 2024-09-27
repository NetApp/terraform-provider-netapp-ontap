
help_and_exit () {
    echo 'This script expects a URL and API_TOKEN to connect to a BD server'
    echo "eg: $0 https://blackduck.domain.com N0TAr3AlKeY@tA1L"
    exit 1
}

# change these 2 values to reflect your project name and version:
export DETECT_PROJECT_NAME="Terraform NetApp ONTAP Provider"
export DETECT_PROJECT_VERSION_NAME=2.0.0
export DETECT_CODE_LOCATION_NAME="${DETECT_PROJECT_NAME}_${DETECT_PROJECT_VERSION_NAME}_code"
export DETECT_BOM_AGGREGATE_NAME="${DETECT_PROJECT_NAME}_${DETECT_PROJECT_VERSION_NAME}_bom"

# set this to true for python or yaml.  false for go or other compiled language
export DETECT_DETECTOR_BUILDLESS=false

# additionally as needed
# detect.detector.search.depth
# see  https://blackducksoftware.github.io/synopsys-detect/latest/  for help

# add go path
#export DETECT_GO_PATH="/usr/bin/go"
# add git path
#export DETECT_GIT_PATH="/usr/bin/git"
# add java path
#export DETECT_JAVA_PATH="/usr/software/java/openjdk-11.0.15_10/bin/java"

if [ -z "$1" ]; then
    help_and_exit
fi

if [ -z "$2" ]; then
    help_and_exit
fi

export BLACKDUCK_URL=$1
export BLACKDUCK_API_TOKEN=$2

bash <(curl -s -L https://detect.synopsys.com/detect7.sh) --blackduck.trust.cert=true