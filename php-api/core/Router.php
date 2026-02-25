<?php
/**
 * 轻量路由器
 */

namespace Core;

class Router
{
    private array $routes = [];
    private array $groupStack = [];

    public function group(string $prefix, callable $callback, array $middleware = []): void
    {
        $this->groupStack[] = [
            'prefix'     => $prefix,
            'middleware'  => $middleware,
        ];
        $callback($this);
        array_pop($this->groupStack);
    }

    public function get(string $path, callable|array $handler, array $middleware = []): void
    {
        $this->addRoute('GET', $path, $handler, $middleware);
    }

    public function post(string $path, callable|array $handler, array $middleware = []): void
    {
        $this->addRoute('POST', $path, $handler, $middleware);
    }

    public function put(string $path, callable|array $handler, array $middleware = []): void
    {
        $this->addRoute('PUT', $path, $handler, $middleware);
    }

    public function delete(string $path, callable|array $handler, array $middleware = []): void
    {
        $this->addRoute('DELETE', $path, $handler, $middleware);
    }

    private function addRoute(string $method, string $path, callable|array $handler, array $middleware): void
    {
        $fullPrefix = '';
        $allMiddleware = [];
        foreach ($this->groupStack as $group) {
            $fullPrefix .= $group['prefix'];
            $allMiddleware = array_merge($allMiddleware, $group['middleware']);
        }
        $allMiddleware = array_merge($allMiddleware, $middleware);

        $fullPath = $fullPrefix . $path;

        $this->routes[] = [
            'method'     => $method,
            'path'       => $fullPath,
            'handler'    => $handler,
            'middleware'  => $allMiddleware,
        ];
    }

    public function dispatch(string $method, string $uri): void
    {
        // 处理 OPTIONS 预检
        if ($method === 'OPTIONS') {
            http_response_code(204);
            exit;
        }

        // 移除 query string
        $uri = strtok($uri, '?');
        // 移除尾部斜杠
        $uri = rtrim($uri, '/') ?: '/';

        foreach ($this->routes as $route) {
            if ($route['method'] !== $method) {
                continue;
            }

            $params = $this->matchRoute($route['path'], $uri);
            if ($params !== false) {
                // 执行中间件
                foreach ($route['middleware'] as $mw) {
                    if (is_string($mw) && class_exists($mw)) {
                        (new $mw())->handle();
                    } elseif (is_callable($mw)) {
                        $mw();
                    }
                }

                // 执行处理器
                $handler = $route['handler'];
                if (is_array($handler)) {
                    [$class, $method] = $handler;
                    (new $class())->$method($params);
                } elseif (is_callable($handler)) {
                    $handler($params);
                }
                return;
            }
        }

        http_response_code(404);
        Response::error(404, '接口不存在');
    }

    /**
     * 路由匹配，支持 :param 参数
     * @return array|false 匹配到的参数数组或 false
     */
    private function matchRoute(string $routePath, string $uri): array|false
    {
        $routeParts = explode('/', trim($routePath, '/'));
        $uriParts   = explode('/', trim($uri, '/'));

        if (count($routeParts) !== count($uriParts)) {
            return false;
        }

        $params = [];
        for ($i = 0; $i < count($routeParts); $i++) {
            if (str_starts_with($routeParts[$i], ':')) {
                $params[substr($routeParts[$i], 1)] = $uriParts[$i];
            } elseif ($routeParts[$i] !== $uriParts[$i]) {
                return false;
            }
        }

        return $params;
    }
}
