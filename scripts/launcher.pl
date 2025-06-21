#!/usr/bin/env perl

use strict;
use warnings;

use Proc::Daemon;
use Getopt::Long qw(:config no_ignore_case bundling);
use Path::Tiny;
use POSIX qw(strftime);

use Data::Dump qw(dump);

my $module         = path(__FILE__);
my $module_name    = $module->basename;
my $main_directory = $module->parent->parent->realpath;

my $executable    = './gin-blog';
my $service_name  = '';
my $command       = '';
my $directory     = '';
my $log_directory = '';
my $quiet         = 0;
my $debug         = 0;
my $exit_code     = 0;

GetOptions(
    "directory|dir"         => \$directory,
    "log_directory|log_dir" => \$log_directory,
    "debug|d"               => \$debug,
    "quiet|q"               => \$quiet
);

if ($debug) {
    print "\@ARGV dmp:\n", dump( \@ARGV ), "\n";
    print "main dir: '" . $main_directory->stringify . "'\n";
}

if ( $directory ne '' ) {
    if ( index( $directory, '/' ) == 0 ) {
        $directory = path($directory)->realpath;
    }
    else {
        $directory = $main_directory->child($directory)->realpath;
    }
}
else {
    $directory = $main_directory;
}

if ( $log_directory ne '' ) {
    if ( index( $log_directory, '/' ) == 0 ) {
        $log_directory = path($log_directory)->realpath;
    }
    else {
        $log_directory = $directory->child($log_directory)->realpath;
    }
}
else {
    $log_directory = path($directory)->child('log');
}

# Create the log directory
if ( $log_directory->can('mkdir') ) {
    $log_directory->mkdir;
}
else {
    $log_directory->mkpath;
}

$service_name = $directory->child($executable)->basename;

$command = $ARGV[0];

if ($debug) {
    print "log dir: '" . $log_directory->stringify . "'\n";
    print "command: '$command'\n";
}

if ( $command eq 'start' ) {
    $exit_code = command_start(
        service_name  => $service_name,
        directory     => $directory,
        log_directory => $log_directory,
        debug         => $debug,
        quiet         => $quiet
    );
}
elsif ( $command eq 'restart' ) {

    # Stop the service
    $exit_code = command_stop(
        service_name  => $service_name,
        log_directory => $log_directory,
        debug         => $debug,
        quiet         => $quiet
    );

    # Start the service again
    $exit_code = command_start(
        service_name  => $service_name,
        directory     => $directory,
        log_directory => $log_directory,
        debug         => $debug,
        quiet         => $quiet
    );
}
elsif ( $command eq 'status' ) {
    $exit_code = command_status(
        service_name  => $service_name,
        log_directory => $log_directory,
        debug         => $debug,
        quiet         => $quiet
    );
}
elsif ( $command eq 'stop' ) {
    $exit_code = command_stop(
        service_name  => $service_name,
        log_directory => $log_directory,
        debug         => $debug,
        quiet         => $quiet
    );
}
else {
    if ( $command ne '' ) {
        if ( !$quiet ) {
            print STDERR "Command '$command': Command not recognized!\n";
        }

        $exit_code = 1;
    }
}

if ( $debug && !$quiet ) {
    print "Script '$module_name': Script finishing with [$exit_code]\n";
}

# Exit with the exit code returned by the command
exit $exit_code;

sub command_start {
    my %params = @_;
    my $log_file =
      $params{log_directory}->child( $params{service_name} . '.log' );

    my $daemon = Proc::Daemon->new(
        work_dir     => $params{directory}->stringify,
        child_STDOUT => $log_file->stringify,
        child_STDERR =>
          $params{log_directory}->child( $params{service_name} . '_error.log' )
          ->stringify,
        pid_file =>
          $params{log_directory}->child( $params{service_name} . '.pid' )
          ->stringify,
        exec_command =>
          $params{directory}->child($executable)->realpath->stringify,
    );

    my $now = strftime "%F %H:%M:%S", localtime;

    $log_file->append(
        $now . " - Service '$params{service_name}': Service launching ...\n" );

    my $pid = $daemon->Init();

    if ( !$params{quiet} ) {
        print
          "Service '$params{service_name}' [PID: '$pid']: Service launched.\n";
    }

    my $check_pid = $daemon->Status($pid);

    $now = strftime "%F %H:%M:%S", localtime;

    if ( $check_pid != 0 ) {
        $log_file->append( $now
              . " - Service '$params{service_name}' [PID: '$check_pid']: Service started.\n"
        );

        if ( !$params{quiet} ) {
            print
"Service '$params{service_name}' [PID: '$check_pid']: Service running.\n";
        }
    }
    else {
        $log_file->append( $now
              . " - Service '$params{service_name}' [PID: '$pid']: Service Start failed!\n"
        );

        if ( !$params{quiet} ) {
            print
"Service '$params{service_name}' [PID: '$pid']: Service Start failed!\n";
        }
    }

    return ( $check_pid != 0 ? 0 : 1 );
}

sub command_status {
    my %params = @_;
    my $pid_file =
      $params{log_directory}->child( $params{service_name} . '.pid' );
    my $pid = get_pid_from_file(
        service_name  => $params{service_name},
        log_directory => $params{log_directory},
        file          => $pid_file,
        debug         => $params{debug},
        quiet         => $params{quiet}
    );
    my $check_pid = 0;

    if ( $pid != 0 ) {
        my $daemon = Proc::Daemon->new( pid_file => $pid_file->stringify );

        if ( $params{debug} && !$params{quiet} ) {
            print "$params{service_name} check pid: '$pid'\n";
        }

        $check_pid = $daemon->Status($pid);

        if ( $check_pid != 0 ) {
            if ( !$params{quiet} ) {
                print
"Service '$params{service_name}' [PID: '$check_pid']: Service running.\n";
            }
        }
        else {
            if ( !$params{quiet} ) {
                print
"Service '$params{service_name}' [PID: '$pid']: Service Run failed!\n";
            }

            # Delete PID file of crashed service
            $pid_file->remove;
        }
    }
    else {
        if ( !$params{quiet} ) {
            if ( $params{debug} ) {
                print "$params{service_name} check pid: '$check_pid'\n";
            }

            print "Service '$params{service_name}': Service not running.\n";
        }

        $check_pid = 1;
    }

    return ( $check_pid != 0 ? 0 : 1 );
}

sub command_stop {
    my %params = @_;
    my $log_file =
      $params{log_directory}->child( $params{service_name} . '.log' );
    my $pid_file =
      $params{log_directory}->child( $params{service_name} . '.pid' );
    my $pid = get_pid_from_file(
        service_name  => $params{service_name},
        log_directory => $params{log_directory},
        file          => $pid_file,
        debug         => $params{debug},
        quiet         => $params{quiet}
    );
    my $proc_count = 0;

    if ( $pid != 0 ) {
        my $daemon = Proc::Daemon->new( pid_file => $pid_file->stringify );

        if ( $params{debug} && !$params{quiet} ) {
            print "$params{service_name} terminate pid: '$pid'\n";
        }

        my $now = strftime "%F %H:%M:%S", localtime;

        $log_file->append( $now
              . " - Service '$params{service_name}' [PID: '$pid']: Service terminating ...\n"
        );

        # Terminate the service process
        $proc_count = $daemon->Kill_Daemon( $pid, 15 );

        $now = strftime "%F %H:%M:%S", localtime;

        $log_file->append( $now
              . " - Service '$params{service_name}' [PID: '$pid']: Service terminated.\n"
        );

        if ( !$params{quiet} ) {
            if ( $proc_count != 0 ) {
                print
"Service '$params{service_name}' [PID: '$pid']: Service terminated.\n";
            }
            else {
                print
"Service '$params{service_name}' [PID: '$pid']: Service not found!\n";
            }
        }

        # Delete the PID file
        $pid_file->remove;
    }
    else {
        if ( !$params{quiet} ) {
            print "Service '$params{service_name}': Service not running.\n";
        }

        $proc_count = 1;
    }

    return ( $proc_count != 0 ? 0 : 1 );
}

sub get_pid_from_file {
    my %params   = @_;
    my $pid_file = $params{file};

    $pid_file //=
      $params{log_directory}->child( $params{service_name} . '.pid' );

    my $pid = eval {
        my $file_slurp = $pid_file->slurp;

        if ( $params{debug} && !$params{quiet} ) {
            print "$params{service_name} file pid: '$file_slurp'\n";
        }

        $file_slurp =~ s/^\s+//;
        $file_slurp =~ s/\s+$//;
        $file_slurp;
    };

    if ($@) {
        $pid = 0;
    }

    return $pid;
}
