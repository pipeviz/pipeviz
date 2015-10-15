<?php

use Drupal\Core\Database\Database;

function pipeviz_drush_command() {
  return array(
    'pipeviz-report' => array(
      'description' => 'Reports the state of the Drupal instance to a pipeviz server',
      'aliases' => array('pvr'),
      'arguments' => array(
        'target' => 'The address of the pipeviz server to which the report should be sent.'
      ),
      'bootstrap' => DRUSH_BOOTSTRAP_MAX,
      'required-arguments' => TRUE,
      'options' => array(
        'db' => 'Report the database configuration.'
      ),
    )
  );
}

// TODO D7 compatibility
function drush_pipeviz_report($target) {
  $message = array();
  if (drush_get_option('db')) {
    $all = Database::getAllConnectionInfo();

    $ls = array();
    foreach ($all as $outer => $sets) {
      foreach ($sets as $name => $info) {
        $conn = array(
          'hostname' => $info['host'],
          'proto' => 'tcp',
        );

        // Add the 3306 by default for mysql.
        // TODO such defaults are clearly handled elsewhere in Drupal's systems; take advantage of them instead of reimplementing here
        if (empty($info['port']) && $info['driver'] == 'mysql') {
          $conn['port'] = 3306;
        }

        $ls[] = array(
          'path' => DRUPAL_ROOT,
          'datasets' => array(
            array(
              'name' => $outer,
              'type' => 'mediated',
              'connNet' => $conn,
              'interaction' => 'rw',
              'subset' => $info['database'],
            ),
          ),
        );
      }
    }

    $message['logic-states'] = $ls;
  } else {
    drush_set_error(dt("No options were passed to pipeviz-report for reporting; no message will be sent."));
    return;
  }

  try {
    $client = new GuzzleHttp\Client();
    $client->request('POST', $target, [
      'body' => json_encode($message),
      ]);
  } catch (GuzzleHttp\Exception\ClientException $e) {
    echo $e->getResponse();
  }
}
