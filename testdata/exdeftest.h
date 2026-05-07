#ifndef EXDEFTEST_H
#define EXDEFTEST_H

#define APP_NAME "ExampleDaemon"
#define APP_VERSION_MAJOR 2
#define APP_VERSION_MINOR 14
#define APP_VERSION_PATCH 7

#define DEFAULT_PORT 8080
#define DEFAULT_TIMEOUT_MS 1500
#define MAX_CLIENTS 256
#define MAX_PATH_LEN 4096
#define CACHE_LINE_SIZE 64

#define FEATURE_LOGGING 1
#define FEATURE_METRICS 1
#define FEATURE_EXPERIMENTAL 0

#define LOG_LEVEL_DEBUG 10
#define LOG_LEVEL_INFO 20
#define LOG_LEVEL_WARN 30
#define LOG_LEVEL_ERROR 40

#define STATUS_OK 0
#define STATUS_RETRY 1
#define STATUS_TIMEOUT 2
#define STATUS_BAD_CONFIG 3

#define ASCII_NUL '\0'
#define ASCII_NEWLINE '\n'
#define ASCII_TAB '\t'
#define ASCII_SPACE ' '
#define ASCII_A 'A'
#define PATH_SEPARATOR '/'

#define U8_MAX ((unsigned char)255)
#define DEFAULT_RETRY_COUNT ((int)3)
#define RETRY_DELAY_SECONDS ((float)1.5)
#define INVALID_HANDLE ((void *)0)

#define FULL_VERSION_MAJOR APP_VERSION_MAJOR
#define FULL_VERSION_MINOR APP_VERSION_MINOR
#define FULL_VERSION_MINORB FULL_VERSION_MINOR
#define DEFAULT_STATUS STATUS_OK
#define DEFAULT_LOG_LEVEL LOG_LEVEL_INFO
#define HTTP_PORT DEFAULT_PORT
#define SOCKET_TIMEOUT_MS DEFAULT_TIMEOUT_MS
#define SOCKET_TIMEOUT_SECONDS (DEFAULT_TIMEOUT_MS / 1000)
#define CONFIG_PATH CONFIG_FILE_NAME
#define API_HEALTH_PATH API_BASE_PATH "/health"

// TODO: support these when ive got bitops
// #define FLAG_READ (1 << 0)
// #define FLAG_WRITE (1 << 1)
// #define FLAG_EXEC (1 << 2)
// #define FLAG_RW (FLAG_READ | FLAG_WRITE)

// NOTE: these won't get processed but keeping to make sure of that lol
#define KB(x) ((x) * 1024)
#define MB(x) (KB(x) * 1024)
#define ARRAY_COUNT(x) (sizeof(x) / sizeof((x)[0]))
#define CLAMP_MIN(x, min) ((x) < (min) ? (min) : (x))

#define API_BASE_PATH "/v1"
#define DEFAULT_HOST "127.0.0.1"
#define CONFIG_FILE_NAME "exampled.conf"

// NOTE: these won't get processed but keeping to make sure of that lol
static int hello() { return 7; }

#endif
