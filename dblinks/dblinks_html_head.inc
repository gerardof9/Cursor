<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>DB Links Desarrollo a Produccion</title>
<style>
  body {
    font-family: system-ui, -apple-system, "Segoe UI", sans-serif;
    font-size: 11px;
    color: #333;
    background: #fafbfc;
    margin: 0;
    padding: 4px 8px;
    line-height: 1.25;
  }

  .report-wrap {
    max-width: 960px;
  }

  .page {
    box-sizing: border-box;
    width: 100%;
    background: #fff;
    padding: 0.5rem 0.6rem;
    border: 1px solid #e8eef3;
    margin-bottom: 0.75rem;
  }

  .report-title {
    box-sizing: border-box;
    width: 100%;
    background: #e8f4fc;
    padding: 0.25rem 0.5rem;
    margin: 0 0 0.35rem;
    border: 1px solid #e8eef3;
    border-bottom: 1px solid #d6e9f5;
    text-align: center;
  }

  .report-title h1 {
    font-size: 11px;
    font-weight: 600;
    margin: 0;
    line-height: 1.2;
    color: #1a4a6e;
  }

  .title-block {
    background: #dceef8;
    padding: 0.25rem 0.5rem;
    margin: 0 0 0.35rem;
    border-bottom: 1px solid #c5dff0;
  }

  .host {
    color: #4a6a82;
    margin: 0;
    font-size: 10px;
    line-height: 1.2;
  }

  .host strong {
    font-weight: 700;
    color: #1a4a6e;
  }

  h2 {
    font-size: 10px;
    font-weight: 600;
    text-transform: none;
    letter-spacing: 0;
    color: #3d6a8a;
    margin: 0.35rem 0 0.2rem;
    padding: 0.15rem 0.35rem;
    background: #eef7fc;
    border-left: 3px solid #b8d9ed;
  }

  h2:first-of-type {
    margin-top: 0;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 0.6rem;
  }

  th {
    text-align: left;
    font-size: 10px;
    font-weight: 600;
    color: #4a6578;
    padding: 0.15rem 0.3rem;
    background: #f0f8fc;
    border-bottom: 1px solid #d6e9f5;
  }

  td {
    padding: 0.15rem 0.3rem;
    vertical-align: top;
    border-bottom: 1px solid #eef2f5;
    line-height: 1.25;
  }

  td:first-child {
    width: 100px;
    font-family: ui-monospace, Consolas, monospace;
    font-size: 10px;
    color: #111;
    background: #fafcfd;
  }

  .list {
    font-family: ui-monospace, Consolas, monospace;
    font-size: 10px;
    color: #444;
    line-height: 1.25;
  }

  .list span:not(:last-child)::after {
    content: ", ";
    color: #999;
  }

  .empty-msg {
    color: #666;
    font-size: 0.9rem;
    margin: 0.5rem 0;
  }
</style>
</head>
<body>

<div class="report-wrap">
<div class="report-title">
  <h1>DB Links Desarrollo a Produccion</h1>
</div>
