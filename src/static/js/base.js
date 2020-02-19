function getCaretPos(id)
{
	let _range = document.getSelection().getRangeAt(0);
	let range = _range.cloneRange();
	range.selectNodeContents(id);

	range.setEnd(_range.endContainer, _range.endOffset);
	return(range.toString().length);
}
