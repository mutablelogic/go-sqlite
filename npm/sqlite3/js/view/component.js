export default class Component {
  static $new(elementName, classNames, ...children) {
    const node = document.createElement(elementName);
    node.replaceChildren(...children);
    classNames.split(' ').forEach((className) => {
      const c = className.trim();
      if (c) {
        node.classList.add(c);
      }
    });
    return node;
  }

  static div(classNames, ...children) {
    return Component.$new('DIV', classNames || '', ...children);
  }

  static badge(classNames, ...children) {
    return Component.$new('SPAN', `badge me-1 ${classNames}`, ...children);
  }

  static span(classNames, ...children) {
    return Component.$new('SPAN', classNames || '', ...children);
  }

  static strong(classNames, ...children) {
    return Component.$new('STRONG', classNames || '', ...children);
  }

  static small(classNames, ...children) {
    return Component.$new('SMALL', classNames || '', ...children);
  }

  static anchor(classNames, href, ...children) {
    const anchor = Component.$new('A', classNames || '', ...children);
    if (href) {
      anchor.setAttribute('href', href);
    }
    return anchor;
  }

  static option(value, name) {
    const node = document.createElement('OPTION');
    node.replaceChildren(name || value);
    node.value = value || '';
    return node;
  }

  static input(classNames, name, value) {
    const input = Component.$new('INPUT', classNames || '');
    input.type = 'text';
    input.name = name || '';
    input.value = value || '';
    return input;
  }
}
