Getting an rpm running:

1. Template and move quadlet files into directory for the rpm build
    - Use make an add as dep for the rpm
    - Need to modify the scripts to template and move but not generate secrets or start
    - Quadlet files need to be templated to handle dynamic config paths
    - Add depend on default.target
2. Modify location of target directories for quadlets config and quadlet systemd files
3. Modify install setup script
    - Secret generation
    - DB scrpi e.g. wait_for and modify role
4. Verify on a fresh VM
5. Document networking commands needed to hit services from outside the VM
6. Publish PR / doc / findings / demo


Questions:
- Upstream / Downstream diffs
    - Is the expectation we will maintain an upstream rpm for this type of installation?
    - How does this rpm relate to the existing ones, should the cli also be bundled?
- Where does the code / spec / scripts for the downstream live?
- What other options might we need NOW, specifically for the user interface or interactions after rpm install?
  - More default options should be added to the application config
  - What other quadlets / system,d config do we need to productionize e.g. health checks?
  - BYO certs support?
  - Auth optionality? (requires more scripting on the install/host side)
  - Custom registry?
  - Move some logic into oneshot containers e.g. db setup?

Script interface
- Quadlet sysyemd output path
- Quadlet configuration output path
- Quadlet configuration host mount path (dir referenced in the .container files)



Sections for doc:

- Outline
- How the rpm works / basic steps
- Document configuration options (right now pws)
- AAP registration expectations
- Alternatives
- Bootc image implementation
- Raise questions


https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html
