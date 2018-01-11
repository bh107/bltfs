FROM centos

COPY ltfssde-2.2.2.0-9802-RHEL7.x86_64.rpm /tmp/ltfssde.rpm

RUN yum -y update
RUN yum -y install vim
RUN yum -y localinstall /tmp/ltfssde.rpm
RUN rm /tmp/ltfssde.rpm
