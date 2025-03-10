Name:           flightctl-installer
Version:        0.5.0
Release:        1%{?dist}
Summary:        Flight Control Installer

License:        Apache-2.0 AND BSD-2-Clause AND BSD-3-Clause AND ISC AND MIT

Source0:        quadlets/
Source1:        preprocess.sh
Source2:        install.sh

BuildArch:      noarch
BuildRequires:  make
Requires:       bash

%description
The flightctl-installer package provides quadlet files and setup for running Flight Control services

%prep
# Run preprocessing script to format the files
sh %{SOURCE1}

%build
# No compilation needed for this package

%install
# Create the target directory
mkdir -p %{buildroot}/opt/flightctl/

# Copy files into the build root
cp -r %{SOURCE0}/quadlets %{buildroot}/opt/flightctl/quadlets
cp %{SOURCE2} %{buildroot}/opt/flightctl

%files
%defattr(0755,root,root,-)
/opt/flightctl/quadlets
/opt/flightctl/install.sh

%post
# Run installation script to move files to the final location
sh /opt/flightctl/install.sh

%changelog
* Mon Mar 10 2025 Dakota Crowder <dcrowder@redhat.com> - 0.0.1
- New specfile for quadlet installer package
