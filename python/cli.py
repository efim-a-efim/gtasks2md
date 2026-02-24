import click
from core import export_tasks, import_tasks

@click.group()
@click.option('--credentials', '-c', type=click.Path(exists=False), help='Path to the OAuth 2.0 credentials.json file.')
@click.pass_context
def cli(ctx, credentials):
    """Google Tasks to Markdown Sync"""
    ctx.ensure_object(dict)
    ctx.obj['CREDENTIALS'] = credentials

@cli.command()
@click.argument('output_path', default='.')
@click.option('--list-name', '-l', help='Specify a single Google Task list name to export (required if output_path is a single file).')
@click.pass_context
def export(ctx, output_path, list_name):
    """Exports task lists from Google Tasks to local Markdown files."""
    credentials = ctx.obj.get('CREDENTIALS')
    try:
        export_tasks(output_path, list_name=list_name, credentials_path=credentials)
    except Exception as e:
        click.echo(f"Error: {str(e)}", err=True)

@cli.command(name='import')
@click.argument('input_path')
@click.option('--list-name', '-l', help='Target Google Tasks list name (optional override).')
@click.pass_context
def import_cmd(ctx, input_path, list_name):
    """Imports task lists from local Markdown files to Google Tasks."""
    credentials = ctx.obj.get('CREDENTIALS')
    try:
        import_tasks(input_path, list_name=list_name, credentials_path=credentials)
    except Exception as e:
        click.echo(f"Error: {str(e)}", err=True)
