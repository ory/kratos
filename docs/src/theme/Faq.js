import React, { useState } from 'react'
import cn from 'classnames'
import styles from './faq.module.css'
//import questions from './faq_questions.json'

const Question = ({ children, tags }) => <div className={tags}>{children}</div>

const TagButton = ({ tag, isSelected, children, toggleSelected }) => (
  <li
    className={cn(
      { [styles.selected]: isSelected },
      tag + '_src-theme-',
      'pills',
      'pills__item',
      { 'styles.pills__item--active': isSelected }
    )}
    onClick={toggleSelected}
  >
    {children}
  </li>
)

const Faq = ({ tags }) => {
  tags = tags.split(' ')
  const [selectedTags, setSelectedTags] = useState(tags)

  const displayFunc = (tags) => {
    for (var i = 0; i < tags.length; i++) {
      if (selectedTags.find((t) => t === tags[i])) {
        return 'block'
      }
    }
    return 'none'
  }

  return (
    <>
      {tags.map((tag) => (
        <TagButton
          key={tag}
          tag={tag}
          isSelected={selectedTags.find((t) => t === tag)}
          toggleSelected={() => {
            if (selectedTags.find((t) => t === tag)) {
              setSelectedTags(selectedTags.filter((t) => t !== tag))
            } else {
              setSelectedTags([...selectedTags, tag])
            }
          }}
        >
          #{tag}
        </TagButton>
      ))}
    </>
  )
}

export { Faq, Question }
